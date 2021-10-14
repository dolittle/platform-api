package staticfiles

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/sirupsen/logrus"
)

type service struct {
	logContext logrus.FieldLogger
	Storage    StorageProxy
	uriPrefix  string
	tenantID   string
}

func NewService(logContext logrus.FieldLogger, uriPrefix string, storage StorageProxy, tenantID string) *service {
	s := &service{
		logContext: logContext,
		uriPrefix:  uriPrefix,
		Storage:    storage,
		tenantID:   tenantID,
	}
	return s
}

func (s *service) ListFiles(w http.ResponseWriter, r *http.Request) {
	containerURL := s.Storage.containerURL
	items := make([]string, 0)
	for marker := (azblob.Marker{}); marker.NotDone(); { // The parens around Marker{} are required to avoid compiler error.
		// Get a result segment starting with the blob indicated by the current Marker.
		listBlob, err := containerURL.ListBlobsFlatSegment(context.TODO(), marker, azblob.ListBlobsSegmentOptions{})
		if err != nil {
			log.Fatal(err)
		}

		marker = listBlob.NextMarker

		for _, blobInfo := range listBlob.Segment.BlobItems {
			name := fmt.Sprintf("/%s/%s",
				strings.Trim(s.uriPrefix, "/"),
				strings.TrimPrefix(blobInfo.Name, s.Storage.defaultPrefix),
			)
			items = append(items, name)
		}
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"files": items,
	})
}

func (s *service) Get(w http.ResponseWriter, r *http.Request) {
	proxy := s.Storage

	key := r.URL.Path
	key = strings.TrimPrefix(key, s.uriPrefix)

	if key[0] == '/' {
		key = key[1:]
	}

	s.logContext.WithFields(logrus.Fields{
		"key": key,
	}).Info("lookup")

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

func (s *service) Upload(w http.ResponseWriter, r *http.Request) {
	proxy := s.Storage
	key := r.URL.Path

	prefix := "/manage/add"
	key = strings.TrimPrefix(key, prefix)

	if key[0] == '/' {
		key = key[1:]
	}

	s.logContext.WithFields(logrus.Fields{
		"key": key,
	}).Info("upload")

	proxy.uploadBlob(w, r, key)
}

func (s *service) Remove(w http.ResponseWriter, r *http.Request) {
	proxy := s.Storage
	key := r.URL.Path

	prefix := "/manage/remove"
	key = strings.TrimPrefix(key, prefix)
	key = strings.TrimPrefix(key, s.uriPrefix)

	if key[0] == '/' {
		key = key[1:]
	}

	s.logContext.WithFields(logrus.Fields{
		"key": key,
	}).Info("remove")
	// TODO we are not in control of the response
	proxy.deleteBlob(w, r, key)
}
