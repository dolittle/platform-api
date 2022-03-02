package studio

import (
	"net/http"

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

	studioConfig, err := s.storageRepo.GetStudioConfig(customerID)

	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, studioConfig)
}
