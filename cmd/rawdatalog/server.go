package rawdatalog

import (
	"log"
	"net/http"
	"time"

	"github.com/dolittle-entropy/platform-api/pkg/middleware"
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

		router := mux.NewRouter()

		stdChainBase := alice.New(middleware.LogTenantUser)
		stdChainWithJSON := stdChainBase.Append(middleware.EnforceJSONHandler)

		natsServer := viper.GetString("rawdatalog.log.nats.server")
		clusterID := viper.GetString("rawdatalog.log.stan.clusterID")
		clientID := viper.GetString("rawdatalog.log.stan.clientID")

		stanConnection := rawdatalog.SetupStan(logrus.WithField("service", "raw-data-log-writer"), natsServer, clusterID, clientID)
		repo := rawdatalog.NewStanLogRepo(stanConnection)
		service := rawdatalog.NewService(logrus.WithField("service", "raw-data-log"), repo)

		router.PathPrefix("/webhook/").Handler(stdChainWithJSON.ThenFunc(service.Webhook)).Methods("POST", "PUT")

		srv := &http.Server{
			Handler:      router,
			Addr:         listenOn,
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}

		log.Fatal(srv.ListenAndServe())
	},
}

func init() {
	RootCmd.AddCommand(serverCMD)
	viper.SetDefault("rawdatalog.server.secret", "change")
	viper.SetDefault("rawdatalog.server.listenOn", "localhost:8080")
	viper.BindEnv("rawdatalog.server.listenOn", "LISTEN_ON")
}
