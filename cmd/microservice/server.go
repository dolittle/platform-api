package microservice

import (
	"log"
	"net/http"
	"time"

	"github.com/dolittle-entropy/platform-api/pkg/middleware"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/application"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice"
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
	Long: `


	fetch('http://localhost:8080/ping').then(d => d.text()).then(d=> console.log(d))
`,
	Run: func(cmd *cobra.Command, args []string) {
		gitRepoBranch := viper.GetString("tools.server.gitRepo.branch")
		if gitRepoBranch == "" {
			panic("GIT_BRANCH required")
		}

		kubeconfig := viper.GetString("tools.server.kubeConfig")
		// TODO hoist localhost into viper
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}

		listenOn := viper.GetString("tools.server.listenOn")
		sharedSecret := viper.GetString("tools.server.secret")

		// create the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		router := mux.NewRouter()

		k8sRepo := platform.NewK8sRepo(clientset, config)
		logrus.SetFormatter(&logrus.JSONFormatter{})
		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			"git@github.com:freshteapot/test-deploy-key.git",
			"/tmp/dolittle-k8s",
			gitRepoBranch,
			// TODO fix this, then update deployment
			"/Users/freshteapot/dolittle/.ssh/test-deploy",
		)

		microserviceService := microservice.NewService(gitRepo, k8sRepo, clientset)
		applicationService := application.NewService(gitRepo, k8sRepo)
		tenantService := tenant.NewService()

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
		stdChain := alice.New(c.Handler, middleware.LogTenantUser, middleware.RestrictHandlerWithHeaderName(sharedSecret, "x-shared-secret"), middleware.EnforceJSONHandler)

		//router.NotFoundHandler = http.HandlerFunc(MyNotFound)

		router.Handle("/microservice", stdChain.ThenFunc(microserviceService.Create)).Methods("POST", "OPTIONS")
		router.Handle("/application", stdChain.ThenFunc(applicationService.Create)).Methods("POST", "OPTIONS")
		router.Handle("/tenant", stdChain.ThenFunc(tenantService.Create)).Methods("POST", "OPTIONS")
		router.Handle("/environment", stdChain.ThenFunc(applicationService.SaveEnvironment)).Methods("POST", "OPTIONS")

		router.Handle("/application/{applicationID}/environment", stdChain.ThenFunc(applicationService.SaveEnvironment)).Methods("POST", "OPTIONS")
		router.Handle("/application/{applicationID}/microservices", stdChain.ThenFunc(microserviceService.GetByApplicationID)).Methods("GET", "OPTIONS")
		router.Handle("/application/{applicationID}", stdChain.ThenFunc(applicationService.GetByID)).Methods("GET", "OPTIONS")
		router.Handle("/applications", stdChain.ThenFunc(applicationService.GetApplications)).Methods("GET", "OPTIONS")

		router.Handle("/application/{applicationID}/environment/{environment}/microservice/{microserviceID}", stdChain.ThenFunc(microserviceService.GetByID)).Methods("GET", "OPTIONS")
		router.Handle("/application/{applicationID}/environment/{environment}/microservice/{microserviceID}", stdChain.ThenFunc(microserviceService.Delete)).Methods("DELETE", "OPTIONS")

		router.Handle("/live/applications", stdChain.ThenFunc(applicationService.GetLiveApplications)).Methods("GET", "OPTIONS")

		router.Handle("/live/application/{applicationID}/microservices", stdChain.ThenFunc(microserviceService.GetLiveByApplicationID)).Methods("GET", "OPTIONS")
		router.Handle("/live/application/{applicationID}/environment/{environment}/microservice/{microserviceID}/podstatus", stdChain.ThenFunc(microserviceService.GetPodStatus)).Methods("GET", "OPTIONS")
		router.Handle("/live/application/{applicationID}/pod/{podName}/logs", stdChain.ThenFunc(microserviceService.GetPodLogs)).Methods("GET", "OPTIONS")

		// kubectl auth can-i list pods --namespace application-11b6cf47-5d9f-438f-8116-0d9828654657 --as be194a45-24b4-4911-9c8d-37125d132b0b --as-group cc3d1c06-ffeb-488c-8b90-a4536c3e6dfa
		router.Handle("/test/can-i", stdChain.ThenFunc(microserviceService.CanI)).Methods("POST")

		// dev-web-adpator.application-{applicationID}.svc.local - kubernetes
		// Lookup service not
		router.Handle("/application/{applicationID}/businessmomentsadaptor/{microserviceID}/save", stdChain.ThenFunc(microserviceService.BusinessMomentsAdaptorSave)).Methods("POST", "OPTIONS")
		router.Handle("/application/{applicationID}/businessmomentsadaptor/{microserviceID}/rawdata", stdChain.ThenFunc(microserviceService.BusinessMomentsAdaptorRawData)).Methods("GET", "OPTIONS")
		router.Handle("/application/{applicationID}/businessmomentsadaptor/{microserviceID}/sync", stdChain.ThenFunc(microserviceService.BusinessMomentsAdaptorSync)).Methods("GET", "OPTIONS")

		srv := &http.Server{
			Handler: router,
			//Addr:         "0.0.0.0:8080",
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

	viper.BindEnv("tools.server.secret", "HEADER_SECRET")
	viper.BindEnv("tools.server.listenOn", "LISTEN_ON")
}
