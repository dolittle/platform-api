package microservice

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dolittle-entropy/platform-api/pkg/middleware"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/application"
	"github.com/dolittle-entropy/platform-api/pkg/platform/businessmoment"
	"github.com/dolittle-entropy/platform-api/pkg/platform/insights"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice"
	"github.com/dolittle-entropy/platform-api/pkg/share"

	gitStorage "github.com/dolittle-entropy/platform-api/pkg/platform/storage/git"
	"github.com/dolittle-entropy/platform-api/pkg/platform/tenant"
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

		gitRepoURL := viper.GetString("tools.server.gitRepo.url")
		if gitRepoURL == "" {
			logContext.WithFields(logrus.Fields{
				"error": "GIT_REPO_URL required",
			}).Fatal("start up")
		}

		gitRepoBranch := viper.GetString("tools.server.gitRepo.branch")
		if gitRepoBranch == "" {
			logContext.WithFields(logrus.Fields{
				"error": "GIT_REPO_BRANCH required",
			}).Fatal("start up")
		}

		gitSshKeysFolder := viper.GetString("tools.server.gitRepo.gitSshKey")
		if gitSshKeysFolder == "" {
			logContext.WithFields(logrus.Fields{
				"error": "GIT_REPO_SSH_KEY required",
			}).Fatal("start up")
		}

		kubeconfig := viper.GetString("tools.server.kubeConfig")
		// TODO hoist localhost into viper
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}

		listenOn := viper.GetString("tools.server.listenOn")
		sharedSecret := viper.GetString("tools.server.secret")
		subscriptionID := viper.GetString("tools.server.azure.subscriptionId")

		// create the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		// Hide secret
		viper.Set("tools.server.secret", fmt.Sprintf("%s***", sharedSecret[:3]))
		logContext.WithFields(logrus.Fields{
			"settings": viper.Get("tools.server"),
		}).Info("start up")

		router := mux.NewRouter()

		k8sRepo := platform.NewK8sRepo(clientset, config)

		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoURL,
			"/tmp/dolittle-k8s",
			gitRepoBranch,
			gitSshKeysFolder,
		)

		microserviceService := microservice.NewService(gitRepo, k8sRepo, clientset)
		applicationService := application.NewService(subscriptionID, gitRepo, k8sRepo)
		tenantService := tenant.NewService()
		businessMomentsService := businessmoment.NewService(logrus.WithField("context", "business-moments-service"), gitRepo, k8sRepo, clientset)
		insightsService := insights.NewService(
			logrus.WithField("context", "insights-service"),
			k8sRepo,
			"query-frontend.system-monitoring-logs.svc.cluster.local:8080",
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
		stdChainBase := alice.New(c.Handler, middleware.LogTenantUser, middleware.RestrictHandlerWithHeaderName(sharedSecret, "x-shared-secret"))
		stdChainWithJSON := stdChainBase.Append(middleware.EnforceJSONHandler)

		//router.NotFoundHandler = http.HandlerFunc(MyNotFound)

		router.Handle("/microservice", stdChainWithJSON.ThenFunc(microserviceService.Create)).Methods("POST", "OPTIONS")
		router.Handle("/application", stdChainWithJSON.ThenFunc(applicationService.Create)).Methods("POST", "OPTIONS")
		router.Handle("/tenant", stdChainWithJSON.ThenFunc(tenantService.Create)).Methods("POST", "OPTIONS")
		router.Handle("/environment", stdChainWithJSON.ThenFunc(applicationService.SaveEnvironment)).Methods("POST", "OPTIONS")

		router.Handle("/application/{applicationID}/environment", stdChainWithJSON.ThenFunc(applicationService.SaveEnvironment)).Methods("POST", "OPTIONS")
		router.Handle("/application/{applicationID}/microservices", stdChainWithJSON.ThenFunc(microserviceService.GetByApplicationID)).Methods("GET", "OPTIONS")
		router.Handle("/application/{applicationID}", stdChainWithJSON.ThenFunc(applicationService.GetByID)).Methods("GET", "OPTIONS")
		router.Handle("/applications", stdChainWithJSON.ThenFunc(applicationService.GetApplications)).Methods("GET", "OPTIONS")
		router.Handle("/application/{applicationID}/personalised-application-info", stdChainWithJSON.ThenFunc(applicationService.GetPersonalisedInfo)).Methods("GET", "OPTIONS")

		router.Handle("/application/{applicationID}/environment/{environment}/microservice/{microserviceID}", stdChainWithJSON.ThenFunc(microserviceService.GetByID)).Methods("GET", "OPTIONS")
		router.Handle("/application/{applicationID}/environment/{environment}/microservice/{microserviceID}", stdChainWithJSON.ThenFunc(microserviceService.Delete)).Methods("DELETE", "OPTIONS")

		router.Handle("/live/applications", stdChainWithJSON.ThenFunc(applicationService.GetLiveApplications)).Methods("GET", "OPTIONS")
		router.Handle("/live/application/{applicationID}/microservices", stdChainWithJSON.ThenFunc(microserviceService.GetLiveByApplicationID)).Methods("GET", "OPTIONS")
		router.Handle("/live/application/{applicationID}/environment/{environment}/microservice/{microserviceID}/podstatus", stdChainWithJSON.ThenFunc(microserviceService.GetPodStatus)).Methods("GET", "OPTIONS")
		router.Handle("/live/application/{applicationID}/pod/{podName}/logs", stdChainBase.ThenFunc(microserviceService.GetPodLogs)).Methods("GET", "OPTIONS")
		router.Handle("/live/application/{applicationID}/configmap/{configMapName}", stdChainBase.ThenFunc(microserviceService.GetConfigMap)).Methods("GET", "OPTIONS")
		router.Handle("/live/application/{applicationID}/secret/{secretName}", stdChainBase.ThenFunc(microserviceService.GetSecret)).Methods("GET", "OPTIONS")

		router.Handle("/live/application/{applicationID}/environment/{environment}/insights/runtime-v1", stdChainWithJSON.ThenFunc(insightsService.GetRuntimeV1)).Methods("GET", "OPTIONS")
		router.Handle("/live/insights/loki/api/v1/query_range", stdChainWithJSON.ThenFunc(insightsService.ProxyLoki)).Methods("GET", "OPTIONS")

		// kubectl auth can-i list pods --namespace application-11b6cf47-5d9f-438f-8116-0d9828654657 --as be194a45-24b4-4911-9c8d-37125d132b0b --as-group cc3d1c06-ffeb-488c-8b90-a4536c3e6dfa
		router.Handle("/test/can-i", stdChainWithJSON.ThenFunc(microserviceService.CanI)).Methods("POST")

		// dev-web-adpator.application-{applicationID}.svc.local - kubernetes
		// Lookup service not
		router.Handle("/application/{applicationID}/environment/{environment}/businessmomentsadaptor/{microserviceID}/save", stdChainWithJSON.ThenFunc(microserviceService.BusinessMomentsAdaptorSave)).Methods("POST", "OPTIONS")
		router.Handle("/application/{applicationID}/environment/{environment}/businessmomentsadaptor/{microserviceID}/rawdata", stdChainWithJSON.ThenFunc(microserviceService.BusinessMomentsAdaptorRawData)).Methods("GET", "OPTIONS")
		router.Handle("/application/{applicationID}/environment/{environment}/businessmomentsadaptor/{microserviceID}/sync", stdChainWithJSON.ThenFunc(microserviceService.BusinessMomentsAdaptorSync)).Methods("GET", "OPTIONS")

		router.Handle(
			"/businessmomententity",
			stdChainWithJSON.ThenFunc(businessMomentsService.SaveEntity),
		).Methods("POST", "OPTIONS")

		router.Handle(
			"/businessmoment",
			stdChainWithJSON.ThenFunc(businessMomentsService.SaveMoment),
		).Methods("POST", "OPTIONS")
		router.Handle(
			"/application/{applicationID}/environment/{environment}/businessmoments",
			stdChainWithJSON.ThenFunc(businessMomentsService.GetMoments),
		).Methods("GET", "OPTIONS")

		router.Handle(
			"/application/{applicationID}/environment/{environment}/businessmoments/microservice/{microserviceID}/entity/{entityID}",
			stdChainWithJSON.ThenFunc(businessMomentsService.DeleteEntity),
		).Methods("DELETE", "OPTIONS")

		router.Handle(
			"/application/{applicationID}/environment/{environment}/businessmoments/microservice/{microserviceID}/moment/{momentID}",
			stdChainWithJSON.ThenFunc(businessMomentsService.DeleteMoment),
		).Methods("DELETE", "OPTIONS")

		// How do I want to load the data? :)

		pathToDB := viper.GetString("tools.server.pathToDb")

		raw, err := ioutil.ReadFile(pathToDB)
		if err != nil {
			panic(err)
		}

		// Make it work, then we can refactor
		repo := share.NewRepoFromJSON(raw)
		logsService := share.NewLogsService(repo)
		router.Handle("/share/logs/latest/by/app/{tenant}/{application}/{environment}", stdChainWithJSON.ThenFunc(logsService.GetLatestByApplication)).Methods("GET", "OPTIONS")
		router.Handle("/share/logs/link", stdChainWithJSON.ThenFunc(logsService.CreateLink)).Methods("POST", "OPTIONS")

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
	serverCMD.Flags().String("kube-config", "", "FullPath to kubeconfig")
	viper.BindPFlag("tools.server.kubeConfig", serverCMD.Flags().Lookup("kube-config"))
	viper.SetDefault("tools.server.secret", "change")
	viper.SetDefault("tools.server.listenOn", "localhost:8080")
	viper.SetDefault("tools.server.azure.subscriptionId", "")

	viper.BindEnv("tools.server.secret", "HEADER_SECRET")
	viper.BindEnv("tools.server.listenOn", "LISTEN_ON")
	viper.BindEnv("tools.server.azure.subscriptionId", "AZURE_SUBSCRIPTION_ID")

	viper.BindEnv("tools.server.pathToDb", "PATH_TO_DB")
}
