package staticfiles

import (
	"net/http"

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

func (s *service) Root(w http.ResponseWriter, r *http.Request) {
	s.Storage.handler(w, r)
}
