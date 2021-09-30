package rawdatalog

import (
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
)

func SetupStan(logContext logrus.FieldLogger, natsServer string, clusterID string, clientID string) stan.Conn {
	opts := []nats.Option{nats.Name("raw-data-log-writer")}
	logContext = logContext.WithFields(logrus.Fields{
		"context":    "raw-data-log-writer",
		"cluster_id": clusterID,
		"client_id":  clientID,
	})

	logContext.Info("Connecting to NATS Server...")
	nc, err := nats.Connect(natsServer, opts...)

	if err != nil {
		panic(err)
	}

	logContext.Info("Connecting to NATS Streaming Server...")
	sc, err := stan.Connect(clusterID, clientID,
		stan.NatsConn(nc),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			logContext.Fatalf("Connection lost, reason: %v", reason)
		}),
		stan.Pings(10, 5),
	)

	if err != nil {
		logContext.Fatalf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, nc.Opts.Url)
	}

	return sc
}
