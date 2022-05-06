package environmentVariables

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type service struct {
	environmentVariablesRepo EnvironmentVariablesRepo
	k8sDolittleRepo          platformK8s.K8sRepo
	logContext               logrus.FieldLogger
}

func NewService(environmentVariablesRepo EnvironmentVariablesRepo, k8sDolittleRepo platformK8s.K8sRepo, logContext logrus.FieldLogger) service {
	return service{
		environmentVariablesRepo: environmentVariablesRepo,
		k8sDolittleRepo:          k8sDolittleRepo,
		logContext:               logContext,
	}
}

func (s *service) GetEnvironmentVariables(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	microserviceID := vars["microserviceID"]
	// TODO lowercase this, would make our lives so much easier
	environment := vars["environment"]

	userID := r.Header.Get("User-ID")
	customerID := r.Header.Get("Tenant-ID")
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)
	if !allowed {
		return
	}

	response := platform.HttpResponseEnvironmentVariables{
		ApplicationID:  applicationID,
		Environment:    environment,
		MicroserviceID: microserviceID,
		Data:           make([]platform.StudioEnvironmentVariable, 0),
	}

	data, err := s.environmentVariablesRepo.GetEnvironmentVariables(applicationID, environment, microserviceID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Data = data
	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (s *service) UpdateEnvironmentVariables(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	applicationID := vars["applicationID"]
	microserviceID := vars["microserviceID"]
	environment := vars["environment"]

	userID := r.Header.Get("User-ID")
	customerID := r.Header.Get("Tenant-ID")

	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)
	if !allowed {
		return
	}

	response := platform.HttpResponseEnvironmentVariables{
		ApplicationID:  applicationID,
		Environment:    environment,
		MicroserviceID: microserviceID,
		Data:           make([]platform.StudioEnvironmentVariable, 0),
	}

	s.logContext.Info("Update environment variables")

	var input platform.HttpResponseEnvironmentVariables
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(b, &input)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	err = s.environmentVariablesRepo.UpdateEnvironmentVariables(applicationID, environment, microserviceID, input.Data)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	data, err := s.environmentVariablesRepo.GetEnvironmentVariables(applicationID, environment, microserviceID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Data = data
	utils.RespondWithJSON(w, http.StatusOK, response)
}
