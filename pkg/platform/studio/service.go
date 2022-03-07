package studio

import (
	"encoding/json"
	"net/http"

	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type service struct {
	storageRepo     storage.Repo
	logContext      logrus.FieldLogger
	roleBindingRepo k8s.RepoRoleBinding
}

func NewService(
	storageRepo storage.Repo,
	logContext logrus.FieldLogger,
	roleBindingRepo k8s.RepoRoleBinding,
) service {
	return service{
		storageRepo:     storageRepo,
		logContext:      logContext,
		roleBindingRepo: roleBindingRepo,
	}
}

// HTTPStudioConfig is the model of data coming from/to Studio. It's different from
// StudioConfig as the properties use camelCasing, which is nicer to use in TypeScript
type HTTPStudioConfig struct {
	BuildOverwrite       bool     `json:"buildOverwrite"`
	DisabledEnvironments []string `json:"disabledEnvironments"`
	CanCreateApplication bool     `json:"canCreateApplication"`
}

func (s *service) Get(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("User-ID")
	vars := mux.Vars(r)
	customerID := vars["customerID"]
	logContext := s.logContext.WithFields(logrus.Fields{
		"customer_id": customerID,
		"method":      "Get",
	})

	hasAccess, err := s.roleBindingRepo.HasUserAdminAccess(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check if user has access")
		return
	}

	if !hasAccess {
		utils.RespondWithError(w, http.StatusForbidden, "You do not have access")
		return
	}

	studioConfig, err := s.storageRepo.GetStudioConfig(customerID)

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to get the studio config")
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpConfig := HTTPStudioConfig{
		BuildOverwrite:       studioConfig.BuildOverwrite,
		DisabledEnvironments: studioConfig.DisabledEnvironments,
		CanCreateApplication: studioConfig.CanCreateApplication,
	}
	utils.RespondWithJSON(w, http.StatusOK, httpConfig)
}

func (s *service) Save(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("User-ID")
	vars := mux.Vars(r)
	customerID := vars["customerID"]
	logContext := s.logContext.WithFields(logrus.Fields{
		"customer_id": customerID,
		"method":      "Get",
	})

	hasAccess, err := s.roleBindingRepo.HasUserAdminAccess(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check if user has access")
		return
	}

	if !hasAccess {
		utils.RespondWithError(w, http.StatusForbidden, "You do not have access")
		return
	}

	var config HTTPStudioConfig
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&config); err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to decode the request body to a studio config")
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	studioConfig := platform.StudioConfig{
		BuildOverwrite:       config.BuildOverwrite,
		DisabledEnvironments: config.DisabledEnvironments,
		CanCreateApplication: config.CanCreateApplication,
	}

	err = s.storageRepo.SaveStudioConfig(customerID, studioConfig)

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to save the studio config")
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondNoContent(w, http.StatusOK)
}
