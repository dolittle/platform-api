package environmentVariables

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type service struct {
	environmentVariablesRepo EnvironmentVariablesRepo
	k8sDolittleRepo          platform.K8sRepo
	logContext               logrus.FieldLogger
}

func NewService(environmentVariablesRepo EnvironmentVariablesRepo, k8sDolittleRepo platform.K8sRepo, logContext logrus.FieldLogger) service {
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
	environment := strings.ToLower(vars["environment"])

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
	s.logContext.Info("Get environment variables")
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
	fmt.Println(vars)
	applicationID := vars["applicationID"]
	microserviceID := vars["microserviceID"]
	environment := strings.ToLower(vars["environment"])

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
	s.logContext.Info("Get environment variables")

	var input platform.HttpResponseEnvironmentVariables
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(b, &input)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	// We are onnly interested in the Data
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
}
