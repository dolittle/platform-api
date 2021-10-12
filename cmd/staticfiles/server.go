package staticfiles

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/dolittle/platform-api/pkg/staticfiles"
	"github.com/gorilla/mux"
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
	go run main.go static-files server
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		// fix: https://github.com/spf13/viper/issues/798
		for _, key := range viper.AllKeys() {
			viper.Set(key, viper.Get(key))
		}

		listenOn := viper.GetString("staticfiles.server.listenOn")
		tenantID := viper.GetString("staticfiles.server.tenantID")
		uriPrefix := viper.GetString("staticfiles.server.uriPrefix")

		azureAccountName := viper.GetString("staticfiles.server.azureStorageName")
		azureAccountKey := viper.GetString("staticfiles.server.azureStorageKey")
		azureBlobContainer := viper.GetString("staticfiles.server.azureBlobContainer")
		azureBlobContainerUriPrefix := viper.GetString("staticfiles.server.azureBlobContainerUriPrefix")

		azureBlobContainerUriPrefix = strings.TrimPrefix(azureBlobContainerUriPrefix, "/")

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

		router.PathPrefix(uriPrefix).HandlerFunc(service.Root).Methods("GET", "POST")
		//router.PathPrefix("/").HandlerFunc(service.Root).Methods("GET", "POST")

		srv := &http.Server{
			Handler:      router,
			Addr:         listenOn,
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}

		logrus.WithField("settings", viper.Get("staticfiles.server")).Info("Starting Server")
		log.Fatal(srv.ListenAndServe())
	},
}

func init() {
	viper.SetDefault("staticfiles.server.listenOn", "localhost:8080")
	viper.SetDefault("staticfiles.server.uriPrefix", "/docs/")
	viper.SetDefault("staticfiles.server.tenantID", "tenant-fake-123")
	viper.SetDefault("staticfiles.server.azureStorageKey", "change-key")
	viper.SetDefault("staticfiles.server.azureStorageName", "change-name")
	viper.SetDefault("staticfiles.server.azureBlobContainer", "change-blob-container")
	viper.SetDefault("staticfiles.server.azureBlobContainerUriPrefix", "/docs/")

	viper.BindEnv("staticfiles.server.listenOn", "LISTEN_ON")
	viper.BindEnv("staticfiles.server.uriPrefix", "URI_PREFIX")
	viper.BindEnv("staticfiles.server.tenantID", "DOLITTLE_TENANT_ID")
	viper.BindEnv("staticfiles.server.azureStorageKey", "AZURE_STORAGE_KEY")
	viper.BindEnv("staticfiles.server.azureStorageName", "AZURE_STORAGE_NAME")
	viper.BindEnv("staticfiles.server.azureBlobContainer", "AZURE_BLOB_CONTAINER")
	viper.BindEnv("staticfiles.server.azureBlobContainerUriPrefix", "AZURE_BLOB_CONTAINER_PREFIX")
}
