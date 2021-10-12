package staticfiles

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/dolittle/platform-api/pkg/middleware"
	"github.com/dolittle/platform-api/pkg/staticfiles"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serverCMD = &cobra.Command{
	Use:   "server",
	Short: "Server",
	Long: `
	
	AZURE_BLOB_CONTAINER="christest" \
	AZURE_STORAGE_NAME="453e04a74f9d42f2b36cd51f" \
	AZURE_STORAGE_KEY="XXX" \
	URI_PREFIX="/doc/" \
	HEADER_SECRET="fake" \
	go run main.go static-files server
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		// fix: https://github.com/spf13/viper/issues/798
		for _, key := range viper.AllKeys() {
			viper.Set(key, viper.Get(key))
		}

		listenOn := viper.GetString("staticfiles.server.listenon")
		tenantID := viper.GetString("staticfiles.server.tenantid")
		uriPrefix := viper.GetString("staticfiles.server.uriprefix")

		azureAccountName := viper.GetString("staticfiles.server.azurestoragename")
		azureAccountKey := viper.GetString("staticfiles.server.azurestoragekey")
		azureBlobContainer := viper.GetString("staticfiles.server.azureblobcontainer")
		azureBlobContainerUriPrefix := viper.GetString("staticfiles.server.azureblobcontaineruriprefix")

		azureBlobContainerUriPrefix = strings.TrimPrefix(azureBlobContainerUriPrefix, "/")
		sharedSecret := viper.GetString("staticfiles.server.sharedsecret")

		router := mux.NewRouter()

		credential, err := azblob.NewSharedKeyCredential(azureAccountName, azureAccountKey)
		if err != nil {
			log.Fatalf("Failed to create shared credentials: %s", err)
		}

		pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})
		azureURL, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", azureAccountName))
		if err != nil {
			log.Fatalf("Failed to create a storage client: %s", err)
		}

		serviceURL := azblob.NewServiceURL(*azureURL, pipeline)
		containerURL := serviceURL.NewContainerURL(azureBlobContainer)

		storageProxy := staticfiles.NewStorageProxy(
			logrus.WithField("service", "storageProxy"),
			&containerURL,
			azureBlobContainerUriPrefix,
		)

		service := staticfiles.NewService(
			logrus.WithField("service", "static-files"),
			uriPrefix,
			*storageProxy,
			tenantID,
		)
		stdChainBase := alice.New(middleware.RestrictHandlerWithSharedSecret(sharedSecret))

		router.Handle(fmt.Sprintf("/%s/list", strings.Trim(uriPrefix, "/")), stdChainBase.ThenFunc(service.ListFiles)).Methods(http.MethodGet, http.MethodOptions)
		router.PathPrefix(uriPrefix).HandlerFunc(service.Get).Methods(http.MethodGet, http.MethodOptions)
		router.PathPrefix(uriPrefix).Handler(stdChainBase.ThenFunc(service.Upload)).Methods(http.MethodPost, http.MethodOptions)

		srv := &http.Server{
			Handler:      router,
			Addr:         listenOn,
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}

		serverSettings := viper.Get("staticfiles.server").(map[string]interface{})
		serverSettings["azurestoragename"] = fmt.Sprintf("%s***", azureAccountName[:3])
		serverSettings["azurestoragekey"] = fmt.Sprintf("%s***", azureAccountKey[:3])
		serverSettings["sharedsecret"] = fmt.Sprintf("%s***", sharedSecret[:3])

		logrus.WithField("settings", viper.Get("staticfiles.server")).Info("Starting Server")
		log.Fatal(srv.ListenAndServe())
	},
}

func init() {
	viper.SetDefault("staticfiles.server.listenon", "localhost:8080")
	viper.SetDefault("staticfiles.server.uriprefix", "/files/")
	viper.SetDefault("staticfiles.server.tenantid", "tenant-fake-123")
	viper.SetDefault("staticfiles.server.azurestoragekey", "change-key")
	viper.SetDefault("staticfiles.server.azurestoragename", "change-name")
	viper.SetDefault("staticfiles.server.azureblobcontainer", "change-blob-container")
	viper.SetDefault("staticfiles.server.azureblobcontaineruriprefix", "/")
	viper.SetDefault("staticfiles.server.sharedsecret", "fake")

	viper.BindEnv("staticfiles.server.listenon", "LISTEN_ON")
	viper.BindEnv("staticfiles.server.uriprefix", "URI_PREFIX")
	viper.BindEnv("staticfiles.server.tenantid", "DOLITTLE_TENANT_ID")
	viper.BindEnv("staticfiles.server.azurestoragekey", "AZURE_STORAGE_KEY")
	viper.BindEnv("staticfiles.server.azurestoragename", "AZURE_STORAGE_NAME")
	viper.BindEnv("staticfiles.server.azureblobcontainer", "AZURE_BLOB_CONTAINER")
	viper.BindEnv("staticfiles.server.azureblobcontaineruriprefix", "AZURE_BLOB_CONTAINER_PREFIX")
	viper.BindEnv("staticfiles.server.sharedsecret", "HEADER_SECRET") // Not a fan of the name, but then we should fix it in the other place :(
}
