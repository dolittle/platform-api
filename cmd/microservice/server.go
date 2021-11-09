package microservice

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dolittle/platform-api/pkg/middleware"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/application"
	"github.com/dolittle/platform-api/pkg/platform/backup"
	"github.com/dolittle/platform-api/pkg/platform/businessmoment"
	"github.com/dolittle/platform-api/pkg/platform/cicd"
	"github.com/dolittle/platform-api/pkg/platform/insights"
	"github.com/dolittle/platform-api/pkg/platform/microservice"
	"github.com/dolittle/platform-api/pkg/platform/microservice/purchaseorderapi"

	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"
	"github.com/dolittle/platform-api/pkg/platform/tenant"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var serverCMD = &cobra.Command{
	Use:   "server",
	Short: "Server to talk to k8s",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logContext := logrus.StandardLogger()
		gitRepoConfig := initGit(logContext)

		// fix: https://github.com/spf13/viper/issues/798
		for _, key := range viper.AllKeys() {
			viper.Set(key, viper.Get(key))
		}

		kubeconfig := viper.GetString("tools.server.kubeConfig")

		getExternalClusterHostFromEnv := false
		if kubeconfig == "incluster" {
			kubeconfig = ""
			getExternalClusterHostFromEnv = true
		}

		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}

		listenOn := viper.GetString("tools.server.listenOn")
		sharedSecret := viper.GetString("tools.server.secret")
		subscriptionID := viper.GetString("tools.server.azure.subscriptionId")

		externalClusterHost := config.Host
		if getExternalClusterHostFromEnv {
			externalClusterHost = viper.GetString("tools.server.kubernetes.externalClusterHost")
		}

		// create the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		// Hide secret
		serverSettings := viper.Get("tools.server").(map[string]interface{})
		serverSettings["secret"] = fmt.Sprintf("%s***", sharedSecret[:3])
		logContext.WithFields(logrus.Fields{
			"settings": viper.Get("tools.server"),
		}).Info("start up")

		router := mux.NewRouter()

		k8sRepo := platform.NewK8sRepo(clientset, config)

		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		microserviceService := microservice.NewService(
			gitRepo,
			k8sRepo,
			clientset,
			logrus.WithField("context", "microservice-service"),
		)
		applicationService := application.NewService(
			subscriptionID,
			clusterHostForDocs,
			gitRepo,
			k8sRepo,
		)
		tenantService := tenant.NewService()
		businessMomentsService := businessmoment.NewService(
			logrus.WithField("context", "business-moments-service"),
			gitRepo,
			k8sRepo,
			clientset,
		)
		insightsService := insights.NewService(
			logrus.WithField("context", "insights-service"),
			k8sRepo,
			"query-frontend.system-monitoring-logs.svc.cluster.local:8080",
		)
		backupService := backup.NewService(
			logrus.WithField("context", "backup-service"),
			gitRepo,
			k8sRepo,
			clientset,
		)
		purchaseorderapiService := purchaseorderapi.NewService(
			gitRepo,
			k8sRepo,
			clientset,
			logrus.WithField("context", "purchase-order-api-service"),
		)

		cicdService := cicd.NewService(
			logrus.WithField("context", "cicd-service"),
			k8sRepo,
		)
		c := cors.New(cors.Options{
			OptionsPassthrough: false,
			Debug:              true,
			// TODO fix this
			AllowedOrigins: []string{"*", "http://localhost:9006"},
			AllowedMethods: []string{
				http.MethodOptions,
				http.MethodHead,
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
			},
			// Not allowing "x-shared-secret" to not allow it to come from the browser
			AllowedHeaders:   []string{"*", "x-secret", "x-tenant"},
			AllowCredentials: true,
		})

		// x-shared-secret not happy with this
		stdChainBase := alice.New(c.Handler, middleware.LogTenantUser, middleware.RestrictHandlerWithSharedSecretAndIDS(sharedSecret, "x-shared-secret"))
		stdChainWithJSON := stdChainBase.Append(middleware.EnforceJSONHandler)

		//router.NotFoundHandler = http.HandlerFunc(MyNotFound)

		router.Handle(
			"/microservice",
			stdChainWithJSON.ThenFunc(microserviceService.Create),
		).Methods(http.MethodPost, http.MethodOptions)
		router.Handle(
			"/microservice",
			stdChainWithJSON.ThenFunc(microserviceService.Update),
		).Methods(http.MethodPut, http.MethodOptions)
		router.Handle(
			"/application",
			stdChainWithJSON.ThenFunc(applicationService.Create),
		).Methods(http.MethodPost, http.MethodOptions)
		router.Handle(
			"/tenant",
			stdChainWithJSON.ThenFunc(tenantService.Create),
		).Methods(http.MethodPost, http.MethodOptions)
		router.Handle(
			"/environment",
			stdChainWithJSON.ThenFunc(applicationService.SaveEnvironment),
		).Methods(http.MethodPost, http.MethodOptions)

		router.Handle(
			"/application/{applicationID}/environment",
			stdChainWithJSON.ThenFunc(applicationService.SaveEnvironment),
		).Methods(http.MethodPost, http.MethodOptions)
		router.Handle(
			"/application/{applicationID}/microservices",
			stdChainWithJSON.ThenFunc(microserviceService.GetByApplicationID),
		).Methods(http.MethodGet, http.MethodOptions)
		router.Handle(
			"/application/{applicationID}",
			stdChainWithJSON.ThenFunc(applicationService.GetByID),
		).Methods(http.MethodGet, http.MethodOptions)
		router.Handle(
			"/applications",
			stdChainWithJSON.ThenFunc(applicationService.GetApplications),
		).Methods(http.MethodGet, http.MethodOptions)
		router.Handle(
			"/application/{applicationID}/personalised-application-info",
			stdChainWithJSON.ThenFunc(applicationService.GetPersonalisedInfo),
		).Methods(http.MethodGet, http.MethodOptions)

		router.Handle(
			"/application/{applicationID}/environment/{environment}/microservice/{microserviceID}",
			stdChainWithJSON.ThenFunc(microserviceService.GetByID),
		).Methods(http.MethodGet, http.MethodOptions)
		router.Handle(
			"/application/{applicationID}/environment/{environment}/microservice/{microserviceID}",
			stdChainWithJSON.ThenFunc(microserviceService.Delete),
		).Methods(http.MethodDelete, http.MethodOptions)

		router.Handle(
			"/live/applications",
			stdChainWithJSON.ThenFunc(applicationService.GetLiveApplications),
		).Methods(http.MethodGet, http.MethodOptions)
		router.Handle(
			"/live/application/{applicationID}/microservices",
			stdChainWithJSON.ThenFunc(microserviceService.GetLiveByApplicationID),
		).Methods(http.MethodGet, http.MethodOptions)
		router.Handle(
			"/live/application/{applicationID}/environment/{environment}/microservice/{microserviceID}/podstatus",
			stdChainWithJSON.ThenFunc(microserviceService.GetPodStatus),
		).Methods(http.MethodGet, http.MethodOptions)
		router.Handle(
			"/live/application/{applicationID}/pod/{podName}/logs",
			stdChainBase.ThenFunc(microserviceService.GetPodLogs),
		).Methods(http.MethodGet, http.MethodOptions)
		router.Handle(
			"/live/application/{applicationID}/configmap/{configMapName}",
			stdChainBase.ThenFunc(microserviceService.GetConfigMap),
		).Methods(http.MethodGet, http.MethodOptions)
		router.Handle(
			"/live/application/{applicationID}/secret/{secretName}", stdChainBase.ThenFunc(microserviceService.GetSecret),
		).Methods(http.MethodGet, http.MethodOptions)

		router.Handle(
			"/live/application/{applicationID}/environment/{environment}/insights/runtime-v1",
			stdChainWithJSON.ThenFunc(insightsService.GetRuntimeV1),
		).Methods(http.MethodGet, http.MethodOptions)
		router.Handle(
			"/live/insights/loki/api/v1/query_range",
			stdChainWithJSON.ThenFunc(insightsService.ProxyLoki),
		).Methods(http.MethodGet, http.MethodOptions)

		// kubectl auth can-i list pods --namespace application-11b6cf47-5d9f-438f-8116-0d9828654657 --as be194a45-24b4-4911-9c8d-37125d132b0b --as-group cc3d1c06-ffeb-488c-8b90-a4536c3e6dfa
		router.Handle("/test/can-i", stdChainWithJSON.ThenFunc(microserviceService.CanI)).Methods(http.MethodPost)

		// dev-web-adpator.application-{applicationID}.svc.local - kubernetes
		// Lookup service not
		router.Handle(
			"/application/{applicationID}/environment/{environment}/businessmomentsadaptor/{microserviceID}/save",
			stdChainWithJSON.ThenFunc(microserviceService.BusinessMomentsAdaptorSave),
		).Methods(http.MethodPost, http.MethodOptions)
		router.Handle(
			"/application/{applicationID}/environment/{environment}/businessmomentsadaptor/{microserviceID}/rawdata",
			stdChainWithJSON.ThenFunc(microserviceService.BusinessMomentsAdaptorRawData),
		).Methods(http.MethodGet, http.MethodOptions)
		router.Handle(
			"/application/{applicationID}/environment/{environment}/businessmomentsadaptor/{microserviceID}/sync",
			stdChainWithJSON.ThenFunc(microserviceService.BusinessMomentsAdaptorSync),
		).Methods(http.MethodGet, http.MethodOptions)

		router.Handle(
			"/businessmomententity",
			stdChainWithJSON.ThenFunc(businessMomentsService.SaveEntity),
		).Methods(http.MethodPost, http.MethodOptions)

		router.Handle(
			"/businessmoment",
			stdChainWithJSON.ThenFunc(businessMomentsService.SaveMoment),
		).Methods(http.MethodPost, http.MethodOptions)
		router.Handle(
			"/application/{applicationID}/environment/{environment}/businessmoments",
			stdChainWithJSON.ThenFunc(businessMomentsService.GetMoments),
		).Methods(http.MethodGet, http.MethodOptions)

		router.Handle(
			"/application/{applicationID}/environment/{environment}/businessmoments/microservice/{microserviceID}/entity/{entityID}",
			stdChainWithJSON.ThenFunc(businessMomentsService.DeleteEntity),
		).Methods(http.MethodDelete, http.MethodOptions)

		router.Handle(
			"/application/{applicationID}/environment/{environment}/businessmoments/microservice/{microserviceID}/moment/{momentID}",
			stdChainWithJSON.ThenFunc(businessMomentsService.DeleteMoment),
		).Methods(http.MethodDelete, http.MethodOptions)

		router.Handle(
			"/backups/logs/latest/by/app/{applicationID}/{environment}",
			stdChainWithJSON.ThenFunc(backupService.GetLatestByApplication),
		).Methods(http.MethodGet, http.MethodOptions)

		router.Handle(
			"/backups/logs/link",
			stdChainWithJSON.ThenFunc(backupService.CreateLink),
		).Methods(http.MethodPost, http.MethodOptions)

		router.Handle(
			"/application/{applicationID}/environment/{environment}/purchaseorderapi/{microserviceID}/datastatus",
			stdChainBase.ThenFunc(purchaseorderapiService.GetDataStatus),
		).Methods(http.MethodGet, http.MethodOptions)

		router.Handle(
			"/application/{applicationID}/cicd/credentials/service-account/devops",
			stdChainBase.ThenFunc(cicdService.GetDevops),
		).Methods(http.MethodGet, http.MethodOptions)

		router.Handle(
			"/application/{applicationID}/cicd/credentials/container-registry",
			stdChainBase.ThenFunc(cicdService.GetContainerRegistryCredentials),
		).Methods(http.MethodGet, http.MethodOptions)

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

	viper.SetDefault("tools.server.secret", "change")
	viper.SetDefault("tools.server.listenOn", "localhost:8080")
	viper.SetDefault("tools.server.azure.subscriptionId", "")
	viper.SetDefault("tools.server.kubernetes.externalClusterHost", "https://cluster-production-three-dns-cf4a27a3.hcp.westeurope.azmk8s.io:443")

	viper.BindEnv("tools.server.secret", "HEADER_SECRET")
	viper.BindEnv("tools.server.listenOn", "LISTEN_ON")
	viper.BindEnv("tools.server.azure.subscriptionId", "AZURE_SUBSCRIPTION_ID")
	viper.BindEnv("tools.server.kubernetes.externalClusterHost", "AZURE_EXTERNAL_CLUSTER_HOST")
}
