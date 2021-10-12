package staticfiles

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/sirupsen/logrus"
)

type StorageProxy struct {
	containerURL  *azblob.ContainerURL
	defaultPrefix string
	logContext    logrus.FieldLogger
}

func NewStorageProxy(logContext logrus.FieldLogger, containerURL *azblob.ContainerURL, defaultPrefix string) *StorageProxy {
	metadataResponse, _ := containerURL.GetProperties(context.Background(), azblob.LeaseAccessConditions{})
	if metadataResponse == nil {
		// TODO this doesnt work, we manually created it
		// Maybe we need to bump the version
		logContext.WithFields(logrus.Fields{
			"containerURL": containerURL,
		}).Info("Creating container")
		containerURL.Create(context.Background(), make(map[string]string), azblob.PublicAccessBlob)
	}
	return &StorageProxy{
		containerURL:  containerURL,
		defaultPrefix: defaultPrefix,
		logContext:    logContext,
	}
}

func (proxy StorageProxy) objectName(name string) string {
	return proxy.defaultPrefix + name
}

func (proxy StorageProxy) handler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path
	if key[0] == '/' {
		key = key[1:]
	}
	fmt.Println(key)
	if r.Method == "GET" {
		proxy.downloadBlob(w, key)
	} else if r.Method == "HEAD" {
		proxy.checkBlobExists(w, key)
	} else if r.Method == "POST" {
		proxy.uploadBlob(w, r, key)
	} else if r.Method == "PUT" {
		proxy.uploadBlob(w, r, key)
	}
}

func (proxy StorageProxy) downloadBlob(w http.ResponseWriter, name string) {
	blockBlobURL := proxy.containerURL.NewBlockBlobURL(proxy.objectName(name))
	get, err := blockBlobURL.Download(context.Background(), 0, 0, azblob.BlobAccessConditions{}, false, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	bufferedReader := bufio.NewReader(get.Body(azblob.RetryReaderOptions{}))
	_, err = bufferedReader.WriteTo(w)
	if err != nil {
		proxy.logContext.WithFields(logrus.Fields{
			"name":  name,
			"error": err,
		}).Error("Failed to serve blob")
	}
}

func (proxy StorageProxy) checkBlobExists(w http.ResponseWriter, name string) {
	blockBlobURL := proxy.containerURL.NewBlockBlobURL(proxy.objectName(name))
	response, err := blockBlobURL.GetProperties(context.Background(), azblob.BlobAccessConditions{}, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(response.StatusCode())
}

func (proxy StorageProxy) uploadBlob(w http.ResponseWriter, r *http.Request, name string) {
	blockBlobURL := proxy.containerURL.NewBlockBlobURL(proxy.objectName(name))

	_, err := azblob.UploadStreamToBlockBlob(
		context.Background(),
		bufio.NewReader(r.Body),
		blockBlobURL,
		azblob.UploadStreamToBlockBlobOptions{},
	)
	if err != nil {
		log.Fatal(err)
	}
	w.WriteHeader(http.StatusCreated)
}

// Make another one to handle file upload
