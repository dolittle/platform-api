package rawdatalog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/sirupsen/logrus"
)

type service struct {
	logContext    logrus.FieldLogger
	repo          Repo
	uriPrefix     string
	topic         string
	tenantID      string
	applicationID string
	environment   string
}

func NewService(logContext logrus.FieldLogger, uriPrefix string, topic string, repo Repo, tenantID string, applicationID string, environment string) service {
	return service{
		logContext:    logContext,
		uriPrefix:     uriPrefix,
		repo:          repo,
		tenantID:      tenantID,
		applicationID: applicationID,
		environment:   environment,
	}
}

func (s *service) Webhook(w http.ResponseWriter, r *http.Request) {
	topic := s.topic
	tenantID := s.tenantID
	applicationID := s.applicationID
	environment := s.environment

	pathname := r.URL.Path
	pathname = strings.TrimPrefix(pathname, s.uriPrefix)
	pathname = strings.TrimPrefix(pathname, "/")
	pathname = strings.TrimSuffix(pathname, "/")
	parts := strings.Split(pathname, "/")

	labels := map[string]string{}
	for index, part := range parts {
		indexStr := fmt.Sprintf("uri-%d", index)
		labels[indexStr] = part
	}

	kind := "TODO"
	metadata := RawMomentMetadata{
		TenantID:      tenantID,
		ApplicationID: applicationID,
		Environment:   environment,
		Labels:        map[string]string{},
	}
	metadata.Labels = labels

	var dst interface{}
	dec := json.NewDecoder(r.Body)

	err := dec.Decode(&dst)
	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error":   err,
			"context": "incoming payload",
		}).Error("webhook")
		utils.RespondWithError(w, http.StatusBadRequest, "Failed to pass payload")
		return
	}

	moment := RawMoment{
		Kind:     kind,
		When:     time.Now().UTC().Unix(),
		Data:     dst,
		Metadata: metadata,
	}

	err = s.repo.Write(topic, moment)
	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error":   err,
			"context": "writing to log",
		}).Error("webhook")
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to write to log")
		return
	}

	utils.RespondNoContent(w, http.StatusOK)
	return
}
