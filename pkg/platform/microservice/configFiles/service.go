package configFiles

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type service struct {
	configFilesRepo ConfigFilesRepo
	k8sDolittleRepo platformK8s.K8sRepo
	logContext      logrus.FieldLogger
}

func NewService(configFilesRepo ConfigFilesRepo, k8sDolittleRepo platformK8s.K8sRepo, logContext logrus.FieldLogger) service {
	return service{
		configFilesRepo: configFilesRepo,
		k8sDolittleRepo: k8sDolittleRepo,
		logContext:      logContext,
	}
}

func (s *service) GetConfigFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	microserviceID := vars["microserviceID"]
	// TODO lowercase this, would make our lives so much easier
	environment := vars["environment"]
	userID := r.Header.Get("User-ID")
	customerID := r.Header.Get("Tenant-ID")

	var data platform.StudioConfigFile

	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)
	if !allowed {
		return
	}

	response := platform.HttpResponseConfigFile{
		ApplicationID:  applicationID,
		Environment:    environment,
		MicroserviceID: microserviceID,
		Data:           data,
	}

	data, err := s.configFilesRepo.GetConfigFile(applicationID, environment, microserviceID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Data = data
	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (s *service) UpdateConfigFiles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	applicationID := vars["applicationID"]
	microserviceID := vars["microserviceID"]
	environment := vars["environment"]

	userID := r.Header.Get("User-ID")
	customerID := r.Header.Get("Tenant-ID")

	r.ParseForm()

	var data platform.StudioConfigFile

	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)

	if !allowed {
		return
	}

	response := platform.HttpResponseConfigFile{
		ApplicationID:  applicationID,
		Environment:    environment,
		MicroserviceID: microserviceID,
		Data:           data,
	}

	s.logContext.Info("Update config files")

	var input platform.HttpResponseConfigFile

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Println(err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	bodyAsString := string(body)

	// We are onnly interested in the Data
	err = s.configFilesRepo.UpdateConfigFiles(applicationID, environment, microserviceID, input.Data)
	if err != nil {
		fmt.Println(err)

		utils.RespondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	data, err = s.configFilesRepo.GetConfigFile(applicationID, environment, microserviceID)
	if err != nil {


		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Data = data
	fmt.Println(bodyAsString)
	utils.RespondWithJSON(w, http.StatusOK, response)
}
