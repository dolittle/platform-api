package insights

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/mongo"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func NewService(logContext logrus.FieldLogger, k8sDolittleRepo platform.K8sRepo) service {
	return service{
		logContext:      logContext,
		k8sDolittleRepo: k8sDolittleRepo,
	}
}

func (s *service) GetRuntimeV1(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	environment := strings.ToLower(vars["environment"])

	userID := r.Header.Get("User-ID")
	tenantID := r.Header.Get("Tenant-ID")
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, tenantID, applicationID, userID)
	if !allowed {
		return
	}

	logContext := s.logContext.WithFields(logrus.Fields{
		"applicationID": applicationID,
		"tenantID":      tenantID,
		"userID":        userID,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoURI := mongo.GetMongoURI(applicationID, environment)

	logContext.WithFields(logrus.Fields{
		"mongoURI": mongoURI,
	}).Info("Connecting to mongo")

	client, err := mongo.SetupMongo(ctx, mongoURI)

	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	dbs := mongo.GetEventStoreDatabases(ctx, client)

	latestEvents := make(map[string]platform.RuntimeLatestEvent, 0)
	latestEventsPerEventType := make(map[string][]platform.RuntimeLatestEvent, 0)
	eventLogCounts := make(map[string]int64, 0)
	runtimeStates := make(map[string][]platform.RuntimeState, 0)

	for _, db := range dbs {
		key := fmt.Sprintf("%s", db)

		_latestEvents, err := mongo.GetLatestEvent(ctx, client, db)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"database": db,
				"error":    err,
				"method":   "mongo.GetLatestEvent",
			}).Info("Skipping db: Failed to get latest event, skipping db")
			continue
		}

		latestEvents[key] = _latestEvents

		_latestEventsPerEventType, err := mongo.GetLatestEventPerEventType(ctx, client, db)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"database": db,
				"error":    err,
				"method":   "mongo.GetLatestEventPerEventType",
			}).Info("Skipping db: Failed to get latest event per event type")
			continue
		}
		latestEventsPerEventType[key] = _latestEventsPerEventType

		eventLogCount, err := mongo.GetEventLogCount(ctx, client, db)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"database": db,
				"error":    err,
				"method":   "mongo.GetEventLogCount",
			}).Info("Skipping db: Failed to get event log count")
			continue
		}
		eventLogCounts[key] = eventLogCount

		_runtimeStates, err := mongo.GetRuntimeStates(ctx, client, db)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"database": db,
				"error":    err,
				"method":   "mongo.GetRuntimeStates",
			}).Info("Skipping db: Failed to get runtime states")
			continue
		}
		runtimeStates[key] = _runtimeStates
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"applicationID":            applicationID,
		"environment":              environment,
		"latestEvents":             latestEvents,
		"latestEventsPerEventType": latestEventsPerEventType,
		"eventLogCounts":           eventLogCounts,
		"runtimeStates":            runtimeStates,
	})
	return
}
