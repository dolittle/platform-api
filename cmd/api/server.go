package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dolittle/platform-api/pkg/git"
	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/middleware"
	"github.com/dolittle/platform-api/pkg/platform/application"
	"github.com/dolittle/platform-api/pkg/platform/backup"
	"github.com/dolittle/platform-api/pkg/platform/businessmoment"
	"github.com/dolittle/platform-api/pkg/platform/cicd"
	"github.com/dolittle/platform-api/pkg/platform/customer"
	"github.com/dolittle/platform-api/pkg/platform/insights"
	"github.com/dolittle/platform-api/pkg/platform/job"
	jobK8s "github.com/dolittle/platform-api/pkg/platform/job/k8s"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice"
	"github.com/dolittle/platform-api/pkg/platform/microservice/environmentVariables"
	"github.com/dolittle/platform-api/pkg/platform/microservice/purchaseorderapi"

	k8sSimple "github.com/dolittle/platform-api/pkg/platform/microservice/simple/k8s"
	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultExternalClusterHost = "externalClusterHost"
)

var serverCMD = &cobra.Command{
	Use:   "server",
	Short: "Server for the api",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logContext := logrus.StandardLogger()
		platformEnvironment := viper.GetString("tools.server.platformEnvironment")
		gitRepoConfig := git.InitGit(logContext, platformEnvironment)

		// fix: https://github.com/spf13/viper/issues/798
		for _, key := range viper.AllKeys() {
			viper.Set(key, viper.Get(key))
		}

		viper.GetViper()
		k8sClient, k8sConfig := platformK8s.InitKubernetesClient()

		externalClusterHost := getExternalClusterHost(
			viper.GetString("tools.server.kubernetes.externalClusterHost"),
			k8sConfig.Host,
		)

		listenOn := viper.GetString("tools.server.listenOn")
		sharedSecret := viper.GetString("tools.server.secret")
		subscriptionID := viper.GetString("tools.server.azure.subscriptionId")
		isProduction := viper.GetBool("tools.server.isProduction")
		// Hide secret
		serverSettings := viper.Get("tools.server").(map[string]interface{})
		serverSettings["secret"] = fmt.Sprintf("%s***", sharedSecret[:3])
		logContext.WithFields(logrus.Fields{
			"settings": viper.Get("tools.server"),
		}).Info("start up")

		router := mux.NewRouter()

		k8sRepo := platformK8s.NewK8sRepo(k8sClient, k8sConfig, logContext.WithField("context", "k8s-repo"))
		k8sRepoV2 := k8s.NewRepo(k8sClient, logContext.WithField("context", "k8s-repo-v2"))

		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		jobResourceConfig := jobK8s.CreateResourceConfigFromViper(viper.GetViper())

		microserviceSimpleRepo := k8sSimple.NewSimpleRepo(k8sClient, k8sRepo, k8sRepoV2, isProduction)

		// TODO I wonder how this works when both are in the same cluster,
		// today via the resources, it is not clear which is which "platform-environment".
		go job.NewCustomerJobListener(k8sClient, gitRepo, logContext.WithField("context", "listener-job-customer"))
		go job.NewApplicationJobListener(k8sClient, gitRepo, logContext.WithField("context", "listener-job-application"))

		microserviceService := microservice.NewService(
			isProduction,
			gitRepo,
			k8sRepo,
			k8sClient,
			microserviceSimpleRepo,
			logrus.WithField("context", "microservice-service"),
		)

		microserviceEnvironmentVariablesService := environmentVariables.NewService(
			environmentVariables.NewEnvironmentVariablesK8sRepo(
				k8sRepo,
				k8sClient,
				logrus.WithField("context", "microservice-environment-variables-repo"),
			),
			k8sRepo,
			logrus.WithField("context", "microservice-environment-variables-service"),
		)

		applicationService := application.NewService(
			subscriptionID,
			externalClusterHost,
			k8sClient,
			gitRepo,
			k8sRepo,
			jobResourceConfig,
			microserviceSimpleRepo,
			logrus.WithField("context", "application-service"),
		)

		customerService := customer.NewService(
			k8sClient,
			gitRepo,
			jobResourceConfig,
			logrus.WithField("context", "customer-service"),
		)
		businessMomentsService := businessmoment.NewService(
			logrus.WithField("context", "business-moments-service"),
			gitRepo,
			k8sRepo,
			k8sClient,
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
			k8sClient,
		)
		purchaseorderapiService := purchaseorderapi.NewService(
			isProduction,
			gitRepo,
			k8sRepo,
			k8sClient,
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
			"/customers",
			stdChainWithJSON.ThenFunc(customerService.GetAll),
		).Methods(http.MethodGet, http.MethodOptions)

		router.Handle(
			"/customer",
			stdChainWithJSON.ThenFunc(customerService.Create),
		).Methods(http.MethodPost, http.MethodOptions)

		router.Handle(
			"/application/{applicationID}/microservices",
			stdChainWithJSON.ThenFunc(microserviceService.GetByApplicationID),
		).Methods(http.MethodGet, http.MethodOptions)
		router.Handle(
			"/application/{applicationID}/check/isonline",
			stdChainWithJSON.ThenFunc(applicationService.IsOnline),
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
			"/live/application/{applicationID}/environment/{environment}/microservice/{microserviceID}/restart",
			stdChainWithJSON.ThenFunc(microserviceService.Restart),
		).Methods(http.MethodDelete, http.MethodOptions)

		router.Handle(
			"/live/application/{applicationID}/environment/{environment}/microservice/{microserviceID}/environment-variables",
			stdChainWithJSON.ThenFunc(microserviceEnvironmentVariablesService.GetEnvironmentVariables),
		).Methods(http.MethodGet, http.MethodOptions)

		router.Handle(
			"/live/application/{applicationID}/environment/{environment}/microservice/{microserviceID}/environment-variables",
			stdChainWithJSON.ThenFunc(microserviceEnvironmentVariablesService.UpdateEnvironmentVariables),
		).Methods(http.MethodPut, http.MethodOptions)

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
	viper.SetDefault("tools.server.isProduction", false)
	viper.SetDefault("tools.server.azure.subscriptionId", "")
	viper.SetDefault("tools.server.kubernetes.externalClusterHost", defaultExternalClusterHost)

	viper.BindEnv("tools.server.secret", "HEADER_SECRET")
	viper.BindEnv("tools.server.listenOn", "LISTEN_ON")
	viper.BindEnv("tools.server.isProduction", "IS_PRODUCTION")
	viper.BindEnv("tools.server.azure.subscriptionId", "AZURE_SUBSCRIPTION_ID")
	viper.BindEnv("tools.server.kubernetes.externalClusterHost", "AZURE_EXTERNAL_CLUSTER_HOST")
}

// getExternalClusterHost Return externalHost if set, otherwise fall back to the internalHost
func getExternalClusterHost(externalHost string, internalHost string) string {
	if externalHost != defaultExternalClusterHost {
		return externalHost
	}
	return internalHost
}