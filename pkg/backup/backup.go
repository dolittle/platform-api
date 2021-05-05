package backup

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/dolittle-entropy/platform-api/pkg/middleware"
	"github.com/dolittle-entropy/platform-api/pkg/share"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/rs/cors"
)

func pong(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func Run(secret string, pathToDB string) {
	raw, err := ioutil.ReadFile(pathToDB)
	if err != nil {
		panic(err)
	}

	repo := share.NewRepoFromJSON(raw)
	logsService := share.NewLogsService(repo)
	pongHandler := http.HandlerFunc(pong)
	c := cors.New(cors.Options{
		OptionsPassthrough: false,
		Debug:              true,
		AllowedOrigins:     []string{"*"},
		AllowedMethods: []string{
			http.MethodOptions,
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"*", "x-secret"},
		AllowCredentials: true,
	})

	stdChain := alice.New(c.Handler, middleware.RestrictHandler(secret), middleware.EnforceJSONHandler)

	router := mux.NewRouter()
	router.Handle("/share/logs/customers", stdChain.ThenFunc(logsService.GetCustomers)).Methods("GET", "OPTIONS")
	router.Handle("/share/logs/applications/{tenant}", stdChain.ThenFunc(logsService.GetApplicationsByTenant)).Methods("GET", "OPTIONS")
	router.Handle("/share/logs/latest/{domain}", stdChain.ThenFunc(logsService.GetLatestByDomain)).Methods("GET", "OPTIONS")
	router.Handle("/share/logs/latest/by/domain/{domain}", stdChain.ThenFunc(logsService.GetLatestByDomain)).Methods("GET", "OPTIONS")
	router.Handle("/share/logs/latest/by/app/{tenant}/{application}/{environment}", stdChain.ThenFunc(logsService.GetLatestByApplication)).Methods("GET", "OPTIONS")
	router.Handle("/share/logs/link", stdChain.ThenFunc(logsService.CreateLink)).Methods("POST", "OPTIONS")
	router.Handle("/ping", pongHandler).Methods("GET")

	srv := &http.Server{
		Handler:      router,
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
