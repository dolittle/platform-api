package rawdatalog

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var readLogsCMD = &cobra.Command{
	Use:   "read-logs",
	Short: "Read events from the log",
	Long: `

	TOPIC=topic.todo \
	STAN_CLIENT_ID=nats-reader \
	STAN_CLUSTER_ID=stan \
	NATS_SERVER=127.0.0.1 \
	go run main.go raw-data-log read-logs
	`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetDefault("topic", "topic.todo")

		viper.BindEnv("topic", "TOPIC")
		viper.BindEnv("rawdatalog.log.stan.clusterID", "STAN_CLUSTER_ID")
		viper.BindEnv("rawdatalog.log.stan.clientID", "STAN_CLIENT_ID")
		viper.BindEnv("rawdatalog.log.nats.server", "NATS_SERVER")

		natsServer := viper.GetString("rawdatalog.log.nats.server")
		clusterID := viper.GetString("rawdatalog.log.stan.clusterID")
		clientID := viper.GetString("rawdatalog.log.stan.clientID")

		topic := viper.GetString("topic")

		logrus.SetFormatter(&logrus.JSONFormatter{})

		opts := []nats.Option{nats.Name("raw-data-log-reader")}
		nc, err := nats.Connect(natsServer, opts...)

		if err != nil {
			panic(err)
		}

		logContext := logrus.WithFields(logrus.Fields{
			"context":    "raw-data-log-reader",
			"cluster_id": clusterID,
			"client_id":  clientID,
		})

		logContext.Info("Connecting to nats server...")
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
		defer logCloser(sc)

		ctx := context.Background()
		ReadOneByOneTilLatest(sc, topic, func(msg *stan.Msg) {
			fmt.Println(string(msg.Data))
		}, false)

		ctx, cancel := context.WithCancel(context.Background())
		latestSubscription := readLatest(ctx, sc, topic, func(msg *stan.Msg) {
			fmt.Println(string(msg.Data))
		})

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		select {
		case <-signals:
			cancel()
		}
		latestSubscription.Unsubscribe()
	},
}

func ReadOneByOneTilLatest(sc stan.Conn, topic string, onRead func(msg *stan.Msg), unsubscribe bool) {
	// Could also use MaxInFlight with the timer to force one by one
	d := 200 * time.Millisecond
	// Initially we shall wait
	ticker := time.NewTicker(500 * time.Millisecond)
	done := make(chan bool)
	handle := func(msg *stan.Msg) {
		ticker.Stop()
		onRead(msg)
		ticker.Reset(d)
	}

	durableName := "reader"
	subscription, _ := sc.Subscribe(
		topic,
		handle,
		stan.DurableName(durableName),
		stan.DeliverAllAvailable(),
		stan.MaxInflight(1),
	)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	select {
	case <-done:
		break
	//case t := <-ticker.C:
	case <-ticker.C:
		break
	}

	if !unsubscribe {
		subscription.Close()
		return
	}

	subscription.Unsubscribe()
	return
}

func readLatest(ctx context.Context, sc stan.Conn, topic string, onRead func(msg *stan.Msg)) stan.Subscription {
	durableName := "reader"
	subscription, _ := sc.Subscribe(
		topic,
		onRead,
		stan.DurableName(durableName),
		stan.MaxInflight(1),
	)
	return subscription
}

func logCloser(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Printf("close error: %s", err)
	}
}
