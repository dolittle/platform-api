package mongo

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SetupMongo
// uri=mongodb://localhost:27018
func GetMongoURI(applicationID string, environment string) string {
	// TODO not hardcode
	return fmt.Sprintf(
		"mongodb://%s-mongo.application-%s.svc.cluster.local:27017",
		environment,
		applicationID,
	)
}

func SetupMongo(ctx context.Context, uri string) (client *mongo.Client, err error) {
	opts := options.ClientOptions{}
	opts.SetDirect(true)
	opts.ApplyURI(uri)

	// Connect to MongoDB
	client, err = mongo.Connect(ctx, &opts)

	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)

	if err != nil {
		return nil, err
	}

	return client, err
}

func GetEventStoreDatabases(ctx context.Context, client *mongo.Client) []string {
	// TODO this is not good enough, due to the fact we allow it to be called whatever is in the json object.
	dbs, _ := client.ListDatabaseNames(ctx, bson.D{})
	filtered := []string{}

	for _, db := range dbs {
		if !strings.Contains(db, "eventstore") {
			continue
		}
		filtered = append(filtered, db)
	}
	return filtered
}

func GetLatestEvent(ctx context.Context, client *mongo.Client, database string) (platform.RuntimeLatestEvent, error) {
	latest := platform.RuntimeLatestEvent{}

	type latestInternal struct {
		Row         primitive.Decimal128 `bson:"row,omitempty"`
		EventTypeId primitive.Binary     `bson:"eventTypeId,omitempty"`
		Occurred    primitive.DateTime   `bson:"occurred,omitempty"`
	}

	c := client.Database(database).Collection("event-log")

	sortStage := bson.D{{"$sort", bson.D{{"_id", -1}}}}
	projectStage := bson.D{
		{
			"$project", bson.D{
				{"row", "$_id"},
				{"eventTypeId", "$Metadata.TypeId"},
				{"occurred", "$Metadata.Occurred"},
			},
		},
	}
	pipeline := mongo.Pipeline{
		sortStage,
		projectStage,
		{
			{"$limit", 1},
		},
	}

	cursor, err := c.Aggregate(ctx, pipeline)

	if err != nil {
		return latest, err
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var internal latestInternal
		if err = cursor.Decode(&internal); err != nil {
			fmt.Println("error", err)
			continue
		}

		aUUID, err := uuid.FromBytes(internal.EventTypeId.Data)
		if err != nil {
			fmt.Println("error", err)
			continue
		}

		latest = platform.RuntimeLatestEvent{
			Row:         internal.Row.String(),
			EventTypeId: aUUID.String(),
			Occurred:    internal.Occurred.Time().Format(time.RFC3339Nano),
		}
		return latest, nil
	}
	return latest, nil
}

func GetLatestEventPerEventType(ctx context.Context, client *mongo.Client, database string) ([]platform.RuntimeLatestEvent, error) {
	latestEvents := []platform.RuntimeLatestEvent{}

	type latestInternal struct {
		Row         primitive.Decimal128 `bson:"row,omitempty"`
		EventTypeId primitive.Binary     `bson:"eventTypeId,omitempty"`
		Occurred    primitive.DateTime   `bson:"occurred,omitempty"`
	}

	c := client.Database(database).Collection("event-log")

	/*
			db.getCollection("event-log").aggregate([
		    {
		        $sort: {
		            _id: -1,
		        }
		     },
		     {
		        $group: {
		            _id: "$Metadata.TypeId",
		            occurred: {$first: "$Metadata.Occurred" },
					row: {$first: "$_id" },
		        }
		    },
		    {
		        $project: {
		            _id: 0,
		            eventTypeId: "$_id",
		            occurred: "$occurred",
					row: "$row"
		        }
		    },
		    {
		        $sort: {
		            occurred: -1,
		        }
		     },
		])
	*/
	sortStage1 := bson.D{{"$sort", bson.D{{"_id", -1}}}}
	groupStage := bson.D{
		{
			"$group", bson.D{
				{"_id", "$Metadata.TypeId"},
				{"occurred", bson.D{{"$first", "$Metadata.Occurred"}}},
				{"row", bson.D{{"$first", "$_id"}}},
			},
		},
	}

	projectStage := bson.D{
		{
			"$project", bson.D{
				{"_id", 0},
				{"eventTypeId", "$_id"},
				{"occurred", "$occurred"},
				{"row", "$row"},
			},
		},
	}
	sortStage2 := bson.D{{"$sort", bson.D{{"occurred", -1}}}}
	pipeline := mongo.Pipeline{
		sortStage1,
		groupStage,
		projectStage,
		sortStage2,
	}

	cursor, err := c.Aggregate(ctx, pipeline)

	if err != nil {
		return latestEvents, err
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var internal latestInternal
		if err = cursor.Decode(&internal); err != nil {
			fmt.Println("error", err)
			continue
		}

		aUUID, err := uuid.FromBytes(internal.EventTypeId.Data)
		if err != nil {
			fmt.Println("error", err)
			continue
		}

		latest := platform.RuntimeLatestEvent{
			Row:         internal.Row.String(),
			EventTypeId: aUUID.String(),
			Occurred:    internal.Occurred.Time().Format(time.RFC3339Nano),
		}
		latestEvents = append(latestEvents, latest)
	}
	return latestEvents, nil
}

// GetEventLogCount
// Return the size of the event log
func GetEventLogCount(ctx context.Context, client *mongo.Client, database string) (int64, error) {
	return client.Database(database).Collection("event-log").CountDocuments(ctx, bson.D{}, &options.CountOptions{})
}

func GetRuntimeStates(ctx context.Context, client *mongo.Client, database string) ([]platform.RuntimeState, error) {
	states := []platform.RuntimeState{}
	type internalState struct {
		Position                  primitive.Decimal128 `bson:"Position,omitempty"`
		EventProcessor            primitive.Binary     `bson:"EventProcessor,omitempty"`
		SourceStream              primitive.Binary     `bson:"SourceStream,omitempty"`
		FailingPartitions         bson.M               `bson:"FailingPartitions,omitempty"`
		LastSuccessfullyProcessed primitive.DateTime   `bson:"LastSuccessfullyProcessed,omitempty"`
	}

	// EventProcessor + SourceStream( 0000) = filter
	// EventProcessor + SourceStream (same) = handler
	// FailingPartitions = empty object

	c := client.Database(database).Collection("stream-processor-states")

	/*
			db.getCollection("stream-processor-states").find({
		    	SourceStream: {
		        	$ne: UUID("00000000-0000-0000-0000-000000000000")
		    	}
			});
	*/
	findOptions := options.Find()
	findOptions.Projection = bson.D{
		{"EventProcessor", 1},
		{"SourceStream", 1},
		{"Position", 1},
		{"FailingPartitions", 1},
		{"LastSuccessfullyProcessed", 1},
	}

	cursor, err := c.Find(ctx, bson.M{
		"SourceStream": bson.M{"$ne": "00000000-0000-0000-0000-000000000000"},
	}, findOptions)

	if err != nil {
		panic(err)
	}

	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var internal internalState
		if err = cursor.Decode(&internal); err != nil {
			log.Fatal(err)
		}

		eventProcessorUUID, err := uuid.FromBytes(internal.EventProcessor.Data)
		if err != nil {
			fmt.Println("error eventProcessorUUID", err)
			continue
		}

		sourceStreamUUID, err := uuid.FromBytes(internal.SourceStream.Data)
		if err != nil {
			fmt.Println("error sourceStreamUUID", err)
			continue
		}

		failingPartitions := make([]platform.RuntimeStateFailingPartition, 0)
		if len(internal.FailingPartitions) >= 0 {
			for partition, info := range internal.FailingPartitions {
				document := info.(primitive.M)
				failingPartitions = append(failingPartitions, platform.RuntimeStateFailingPartition{
					LastFailed:         document["LastFailed"].(primitive.DateTime).Time().String(),
					Partition:          partition,
					Position:           document["Position"].(primitive.Decimal128).String(),
					ProcessingAttempts: document["ProcessingAttempts"].(int32),
					Reason:             document["Reason"].(string),
					RetryTime:          document["RetryTime"].(primitive.DateTime).Time().String(),
				})
			}

		}

		states = append(states, platform.RuntimeState{
			Position:                  internal.Position.String(),
			EventProcessor:            eventProcessorUUID.String(),
			SourceStream:              sourceStreamUUID.String(),
			LastSuccessfullyProcessed: internal.LastSuccessfullyProcessed.Time().String(),
			FailingPartitions:         failingPartitions,
		})
	}
	return states, nil
}
