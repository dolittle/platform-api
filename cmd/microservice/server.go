package microservice

import (
	"log"
	"net/http"
	"time"

	"github.com/dolittle-entropy/platform-api/pkg/platform/application"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice"
	"github.com/dolittle-entropy/platform-api/pkg/platform/tenant"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/rs/cors"
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
		kubeconfig := viper.GetString("tools.server.kubeConfig")
		// TODO hoist localhost into viper
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}

		// create the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		router := mux.NewRouter()

		microserviceService := microservice.NewService(clientset)
		applicationService := application.NewService(clientset)
		tenantService := tenant.NewService()

		c := cors.New(cors.Options{
			OptionsPassthrough: false,
			Debug:              true,
			AllowedOrigins:     []string{"*", "http://localhost:9006"},
			AllowedMethods: []string{
				http.MethodOptions,
				http.MethodHead,
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
			},
			AllowedHeaders:   []string{"*", "x-secret", "x-tenant"},
			AllowCredentials: true,
		})

		stdChain := alice.New(c.Handler)

		router.Handle("/microservice", stdChain.ThenFunc(microserviceService.Create)).Methods("POST", "OPTIONS")
		router.Handle("/application", stdChain.ThenFunc(applicationService.Create)).Methods("POST", "OPTIONS")
		router.Handle("/tenant", stdChain.ThenFunc(tenantService.Create)).Methods("POST", "OPTIONS")

		router.Handle("/application/{applicationID}/environment", stdChain.ThenFunc(applicationService.SaveEnvironment)).Methods("POST", "OPTIONS")
		router.Handle("/application/{applicationID}/microservices", stdChain.ThenFunc(microserviceService.GetByApplicationID)).Methods("GET", "OPTIONS")

		router.Handle("/application/{applicationID}/microservice/{microserviceID}", stdChain.ThenFunc(microserviceService.GetByID)).Methods("GET", "OPTIONS")
		router.Handle("/application/{applicationID}/microservice/{microserviceID}", stdChain.ThenFunc(microserviceService.Delete)).Methods("DELETE", "OPTIONS")

		router.Handle("/live/tenant/{tenantID}/applications", stdChain.ThenFunc(applicationService.GetLiveApplications)).Methods("GET", "OPTIONS")
		router.Handle("/live/application/{applicationID}/microservices", stdChain.ThenFunc(microserviceService.GetLiveByApplicationID)).Methods("GET", "OPTIONS")
		//router.Handle("/live/application/{applicationID}/microservice/{microserviceID}", stdChain.ThenFunc(microserviceService.GetLiveByID)).Methods("GET", "OPTIONS")
		router.Handle("/live/application/{applicationID}/microservice/{microserviceID}/podstatus/{environment}", stdChain.ThenFunc(microserviceService.GetPodStatus)).Methods("GET", "OPTIONS")
		router.Handle("/live/application/{applicationID}/pod/{podName}/logs", stdChain.ThenFunc(microserviceService.GetPodLogs)).Methods("GET", "OPTIONS")

		srv := &http.Server{
			Handler: router,
			//Addr:         "0.0.0.0:8080",
			Addr:         "localhost:8080",
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

}
