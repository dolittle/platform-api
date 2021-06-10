package rawdatalog

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dolittle-entropy/platform-api/pkg/rawdatalog"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serverCMD = &cobra.Command{
	Use:   "server",
	Short: "Server",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		listenOn := viper.GetString("rawdatalog.server.listenOn")
		webhookRepoType := strings.ToLower(viper.GetString("rawdatalog.server.webhookRepo"))
		webhookUriPrefix := strings.ToLower(viper.GetString("rawdatalog.server.webhookUriPrefix"))
		pathToMicroserviceConfig := viper.GetString("rawdatalog.server.microserviceConfig")
		tenantID := viper.GetString("rawdatalog.server.tenantID")
		applicationID := viper.GetString("rawdatalog.server.applicationID")
		environment := viper.GetString("rawdatalog.server.environment")
		topic := viper.GetString("rawdatalog.log.topic")

		router := mux.NewRouter()

		// Not needed, but maybe we want some middlewares?
		// Secret lookup could be 1
		stdChain := alice.New()

		var repo rawdatalog.Repo
		switch webhookRepoType {
		case "stdout":
			repo = rawdatalog.NewStdoutLogRepo()
		case "nats":
			natsServer := viper.GetString("rawdatalog.log.nats.server")
			clusterID := viper.GetString("rawdatalog.log.stan.clusterID")
			clientID := viper.GetString("rawdatalog.log.stan.clientID")

			stanConnection := rawdatalog.SetupStan(logrus.WithField("service", "raw-data-log-writer"), natsServer, clusterID, clientID)
			repo = rawdatalog.NewStanLogRepo(stanConnection)
		default:
			panic(fmt.Sprintf("WEBHOOK_REPO %s not supported, pick stdout or nats", webhookRepoType))
		}

		service := rawdatalog.NewService(
			logrus.WithField("service", "raw-data-log"),
			webhookUriPrefix,
			pathToMicroserviceConfig,
			topic,
			repo,
			tenantID,
			applicationID,
			environment,
		)
		router.PathPrefix(webhookUriPrefix).Handler(stdChain.ThenFunc(service.Webhook)).Methods("POST", "PUT")

		srv := &http.Server{
			Handler:      router,
			Addr:         listenOn,
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}

		logrus.WithField("settings", viper.AllSettings()).Info("Starting Server")
		log.Fatal(srv.ListenAndServe())
	},
}

func init() {
	RootCmd.AddCommand(serverCMD)
	viper.SetDefault("rawdatalog.server.secret", "change")
	viper.SetDefault("rawdatalog.server.listenOn", "localhost:8080")
	viper.SetDefault("rawdatalog.server.webhookRepo", "stdout")
	viper.SetDefault("rawdatalog.server.webhookUriPrefix", "/webhook/")
	viper.SetDefault("rawdatalog.server.microserviceConfig", "/tmp/ms.json")
	viper.SetDefault("rawdatalog.server.tenantID", "tenant-fake-123")
	viper.SetDefault("rawdatalog.server.applicationID", "application-fake-123")
	viper.SetDefault("rawdatalog.server.environment", "environment-fake-123")

	viper.SetDefault("rawdatalog.log.topic", "topic.todo")
	viper.SetDefault("rawdatalog.log.stan.clusterID", "stan")
	viper.SetDefault("rawdatalog.log.stan.clientID", "webhook-inserter")
	viper.SetDefault("rawdatalog.log.nats.server", "127.0.0.1")

	viper.BindEnv("rawdatalog.server.listenOn", "LISTEN_ON")
	viper.BindEnv("rawdatalog.server.webhookRepo", "WEBHOOK_REPO")
	viper.BindEnv("rawdatalog.server.microserviceConfig", "MICROSERVICE_CONFIG")
	viper.BindEnv("rawdatalog.server.webhookUriPrefix", "WEBHOOK_PREFIX")
	viper.BindEnv("rawdatalog.server.tenantID", "DOLITTLE_TENANT_ID")
	viper.BindEnv("rawdatalog.server.applicationID", "DOLITTLE_APPLICATION_ID")
	viper.BindEnv("rawdatalog.server.environment", "DOLITTLE_ENVIRONMENT")

	viper.BindEnv("rawdatalog.log.stan.clusterID", "STAN_CLUSTER_ID")
	viper.BindEnv("rawdatalog.log.stan.clientID", "STAN_CLIENT_ID")
	viper.BindEnv("rawdatalog.log.nats.server", "NATS_SERVER")
	viper.BindEnv("rawdatalog.log.topic", "TOPIC")

}
