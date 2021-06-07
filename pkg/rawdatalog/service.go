package rawdatalog

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/sirupsen/logrus"
)

type service struct {
	logContext logrus.FieldLogger
	repo       Repo
}

func NewService(logContext logrus.FieldLogger, repo Repo) service {
	return service{
		logContext: logContext,
		repo:       repo,
	}
}

func (s *service) Webhook(w http.ResponseWriter, r *http.Request) {
	topic := "topic.todo"
	tenantID := "TODO"
	applicationID := "TODO"
	environment := "TODO"
	kind := "TODO"
	metadata := RawMomentMetadata{
		TenantID:      tenantID,
		ApplicationID: applicationID,
		Environment:   environment,
	}

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

	data, _ := json.Marshal(dst)
	moment := RawMoment{
		Kind:     kind,
		When:     time.Now().UTC().Unix(),
		Data:     data,
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
