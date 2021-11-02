package staticfiles

import (
	"net/http"
	"strings"

	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/sirupsen/logrus"
)

type service struct {
	logContext logrus.FieldLogger
	storage    StorageProxy
	uriPrefix  string
	tenantID   string
}

func NewService(logContext logrus.FieldLogger, uriPrefix string, storage StorageProxy, tenantID string) *service {
	s := &service{
		logContext: logContext,
		uriPrefix:  uriPrefix,
		storage:    storage,
		tenantID:   tenantID,
	}
	return s
}

func (s *service) ListFiles(w http.ResponseWriter, r *http.Request) {
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

func (s *service) Get(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path
	key = strings.TrimPrefix(key, s.uriPrefix)

	if key[0] == '/' {
		key = key[1:]
	}

	s.logContext.WithFields(logrus.Fields{
		"key": key,
	}).Info("lookup")

	s.storage.Download(w, key)
}

func (s *service) Upload(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path

	prefix := "/manage/add"
	key = strings.TrimPrefix(key, prefix)

	if key[0] == '/' {
		key = key[1:]
	}

	s.logContext.WithFields(logrus.Fields{
		"key": key,
	}).Info("upload")

	s.storage.Upload(w, r, key)
}

func (s *service) Remove(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path

	prefix := "/manage/remove"
	key = strings.TrimPrefix(key, prefix)
	// prefix = server listening
	// uriPrefix comes next as its included in the path from the frontend
	key = strings.TrimPrefix(key, s.uriPrefix)

	if key[0] == '/' {
		key = key[1:]
	}

	s.logContext.WithFields(logrus.Fields{
		"key": key,
	}).Info("remove")
	// TODO we are not in control of the response
	s.storage.Delete(w, r, key)
}
