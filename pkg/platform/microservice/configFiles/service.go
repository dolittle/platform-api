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

func (s *service) GetConfigFilesNamesList(w http.ResponseWriter, r *http.Request) {
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

	response := platform.HttpResponseConfigFilesNamesList{
		ApplicationID:  applicationID,
		Environment:    environment,
		MicroserviceID: microserviceID,
		Data:           []string{},
	}

	data, err := s.configFilesRepo.GetConfigFilesNamesList(applicationID, environment, microserviceID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, name := range data {
		response.Data = append(response.Data, name)
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

	file, handler, err := r.FormFile("file")


	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)

	if !allowed {
		return
	}

	response := platform.HttpResponseConfigFilesNamesList{
		ApplicationID:  applicationID,
		Environment:    environment,
		MicroserviceID: microserviceID,
	}

	s.logContext.Info("Update config files")

	var input platform.StudioConfigFile

	body, err := ioutil.ReadAll(file)

	if err != nil {
		fmt.Println(err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	bodyAsString := string(body)

	input.BinaryData = body
	input.Name = handler.Filename

	// We are onnly interested in the Data
	err = s.configFilesRepo.UpdateConfigFiles(applicationID, environment, microserviceID, input)
	if err != nil {
		fmt.Println(err)
		utils.RespondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	fmt.Println(bodyAsString)
	utils.RespondWithJSON(w, http.StatusOK, response)
}
