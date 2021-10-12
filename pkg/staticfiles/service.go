package staticfiles

import (
	"net/http"
	"strings"

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
	key = strings.TrimPrefix(key, s.uriPrefix)

	if key[0] == '/' {
		key = key[1:]
	}

	s.logContext.WithFields(logrus.Fields{
		"key": key,
	}).Info("upload")

	proxy.uploadBlob(w, r, key)
}
