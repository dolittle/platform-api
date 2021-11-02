package staticfiles

import (
	"net/http"
	"strings"

	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/sirupsen/logrus"
)

type Service struct {
	logContext logrus.FieldLogger
	storage    StorageProxy
	uriPrefix  string
	tenantID   string
}

func NewService(logContext logrus.FieldLogger, uriPrefix string, storage StorageProxy, tenantID string) Service {
	s := Service{
		logContext: logContext,
		uriPrefix:  uriPrefix,
		storage:    storage,
		tenantID:   tenantID,
	}
	return s
}

func (s Service) ListFiles(w http.ResponseWriter, r *http.Request) {
	items, err := s.storage.ListFiles(s.uriPrefix)
	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error":  err,
			"method": "service.ListFiles",
		}).Error("Listing files")
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get items")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"files": items,
	})
}

func (s Service) Get(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path
	key = strings.TrimPrefix(key, s.uriPrefix)
	s.logContext.WithFields(logrus.Fields{
		"key": key,
	}).Info("lookup")
	s.storage.Download(w, key)
}

func (s Service) Upload(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path

	prefix := "/manage/add"
	key = strings.TrimPrefix(key, prefix)
	// TODO We are not including the uriSuffix in the upload, maybe we shoul
	key = strings.TrimPrefix(key, "/")
	s.logContext.WithFields(logrus.Fields{
		"key": key,
	}).Info("upload")

	s.storage.Upload(w, r, key)
}

func (s Service) Remove(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path

	prefix := "/manage/remove"
	key = strings.TrimPrefix(key, prefix)
	// prefix = server listening
	// uriPrefix comes next as its included in the path from the frontend
	key = strings.TrimPrefix(key, s.uriPrefix)

	s.logContext.WithFields(logrus.Fields{
		"key": key,
	}).Info("remove")
	// TODO we are not in control of the response
	s.storage.Delete(w, r, key)
}
