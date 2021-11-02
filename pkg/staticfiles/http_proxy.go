package staticfiles

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/sirupsen/logrus"
)

type azureStorageProxy struct {
	containerURL  *azblob.ContainerURL
	defaultPrefix string
	logContext    logrus.FieldLogger
}

type StorageProxy interface {
	Download(w http.ResponseWriter, name string)
	Upload(w http.ResponseWriter, r *http.Request, name string)
	Delete(w http.ResponseWriter, r *http.Request, name string)
	Exists(w http.ResponseWriter, name string)
	ListFiles(uriPrefix string) ([]string, error)
}

func NewStorageProxy(logContext logrus.FieldLogger, containerURL *azblob.ContainerURL, defaultPrefix string) azureStorageProxy {
	metadataResponse, _ := containerURL.GetProperties(context.Background(), azblob.LeaseAccessConditions{})
	if metadataResponse == nil {
		// TODO this doesnt work, we manually created it
		// Maybe we need to bump the version
		logContext.WithFields(logrus.Fields{
			"containerURL": containerURL,
		}).Info("Creating container")
		containerURL.Create(context.Background(), make(map[string]string), azblob.PublicAccessBlob)
	}

	return azureStorageProxy{
		containerURL:  containerURL,
		defaultPrefix: defaultPrefix,
		logContext:    logContext,
	}
}

func (proxy azureStorageProxy) objectName(name string) string {
	return proxy.defaultPrefix + name
}

func (proxy azureStorageProxy) ListFiles(uriPrefix string) ([]string, error) {
	containerURL := proxy.containerURL
	files := make([]string, 0)
	for marker := (azblob.Marker{}); marker.NotDone(); { // The parens around Marker{} are required to avoid compiler error.
		// Get a result segment starting with the blob indicated by the current Marker.
		listBlob, err := containerURL.ListBlobsFlatSegment(context.TODO(), marker, azblob.ListBlobsSegmentOptions{})
		if err != nil {
			return files, err
		}

		marker = listBlob.NextMarker

		for _, blobInfo := range listBlob.Segment.BlobItems {
			name := fmt.Sprintf("/%s/%s",
				strings.Trim(uriPrefix, "/"),
				strings.TrimPrefix(blobInfo.Name, proxy.defaultPrefix),
			)
			files = append(files, name)
		}
	}
	return files, nil
}

func (proxy azureStorageProxy) Download(w http.ResponseWriter, name string) {
	blockBlobURL := proxy.containerURL.NewBlockBlobURL(proxy.objectName(name))
	get, err := blockBlobURL.Download(context.Background(), 0, 0, azblob.BlobAccessConditions{}, false, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// TODO we could pass in the desire to trigger Content-Disposition
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Disposition
	bufferedReader := bufio.NewReader(get.Body(azblob.RetryReaderOptions{}))
	_, err = bufferedReader.WriteTo(w)
	if err != nil {
		proxy.logContext.WithFields(logrus.Fields{
			"method": "Download",
			"name":   name,
			"error":  err,
		}).Error("Failed to Download")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (proxy azureStorageProxy) Exists(w http.ResponseWriter, name string) {
	blockBlobURL := proxy.containerURL.NewBlockBlobURL(proxy.objectName(name))
	response, err := blockBlobURL.GetProperties(context.Background(), azblob.BlobAccessConditions{}, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(response.StatusCode())
}

func (proxy azureStorageProxy) Upload(w http.ResponseWriter, r *http.Request, name string) {
	blockBlobURL := proxy.containerURL.NewBlockBlobURL(proxy.objectName(name))

	_, err := azblob.UploadStreamToBlockBlob(
		context.Background(),
		bufio.NewReader(r.Body),
		blockBlobURL,
		azblob.UploadStreamToBlockBlobOptions{},
	)
	if err != nil {
		// TODO perhaps this should not be fatal
		proxy.logContext.WithFields(logrus.Fields{
			"method": "Upload",
			"error":  err,
			"name":   name,
		}).Error("issue uploading")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (proxy azureStorageProxy) Delete(w http.ResponseWriter, r *http.Request, name string) {
	blockBlobURL := proxy.containerURL.NewBlockBlobURL(proxy.objectName(name))

	_, err := blockBlobURL.Delete(r.Context(), azblob.DeleteSnapshotsOptionNone, azblob.BlobAccessConditions{})
	if err != nil {
		proxy.logContext.WithFields(logrus.Fields{
			"method": "Delete",
			"error":  err,
			"name":   name,
		}).Error("issue deleting")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
