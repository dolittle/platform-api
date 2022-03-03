package studio

import (
	"encoding/json"
	"net/http"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type service struct {
	storageRepo storage.Repo
	logContext  logrus.FieldLogger
}

func NewService(
	storageRepo storage.Repo,
	logContext logrus.FieldLogger,
) service {
	return service{
		storageRepo: storageRepo,
		logContext:  logContext,
	}
}

func (s *service) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerID := vars["customerID"]
	logContext := s.logContext.WithFields(logrus.Fields{
		"customer_id": customerID,
		"method":      "Get",
	})

	studioConfig, err := s.storageRepo.GetStudioConfig(customerID)

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to get the studio config")
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, studioConfig)
}

func (s *service) Save(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerID := vars["customerID"]
	logContext := s.logContext.WithFields(logrus.Fields{
		"customer_id": customerID,
		"method":      "Get",
	})

	var config platform.StudioConfig
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&config); err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to decode the request body to a studio config")
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := s.storageRepo.SaveStudioConfig(customerID, config)

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to save the studio config")
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondNoContent(w, http.StatusOK)
}
