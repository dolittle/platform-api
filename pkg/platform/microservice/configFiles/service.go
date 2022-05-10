package configFiles

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

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

	logContext := s.logContext.WithFields(logrus.Fields{
		"method":          "GetConfigFilesNamesList",
		"application_id":  applicationID,
		"microservice_id": microserviceID,
		"environment":     environment,
	})
	
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)
	if !allowed {
		logContext.Info("UpdateConfigFiles: not allowed ")

		return
	}
	
	data, err := s.configFilesRepo.GetConfigFilesNamesList(applicationID, environment, microserviceID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	response := platform.HttpResponseConfigFilesNamesList{
		ApplicationID:  applicationID,
		Environment:    environment,
		MicroserviceID: microserviceID,
		Data:           data,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (s *service) UpdateConfigFiles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	applicationID := vars["applicationID"]
	microserviceID := vars["microserviceID"]
	environment := vars["environment"]

	userID := r.Header.Get("User-ID")
	customerID := r.Header.Get("Tenant-ID")
	
	logContext := s.logContext.WithFields(logrus.Fields{
		"method":          "UpdateConfigFiles",
		"application_id":  applicationID,
		"microservice_id": microserviceID,
		"environment":     environment,
	})

	r.ParseForm()

	file, handler, err := r.FormFile("file")

	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)

	if !allowed {
		logContext.Info("UpdateConfigFiles: not allowed ")

		return
	}

	if file == nil {
		msg := "UpdateConfigFiles ERROR: No file"

		logContext.Info(msg)

		utils.RespondWithError(w, http.StatusBadRequest, msg)
		return
	}

	if strings.TrimSpace(handler.Filename) != handler.Filename {
		msg := "UpdateEnvironmentVariables ERROR: No spaces allowed in config file name"

		logContext.Info(msg)

		utils.RespondWithError(w, http.StatusBadRequest, msg)
		return
	}

	// file size limit from header.Size()
	if handler.Size > 3145728 {
		msg := "UpdateConfigFiles ERROR: File size too large"

		logContext.Info(msg)

		utils.RespondWithError(w, http.StatusBadRequest, msg)
		return
	}

	s.logContext.Info("Update config files")

	
	body, err := ioutil.ReadAll(file)
	
	if err != nil {
		msg := "UpdateConfigFiles ERROR: Invalid file"

		logContext.Info(msg)

		utils.RespondWithError(w, http.StatusBadRequest, "Invalid file")
		return
	}
	defer r.Body.Close()
	
	var input MicroserviceConfigFile
	input.BinaryData = body
	input.Name = handler.Filename

	err = s.configFilesRepo.AddEntryToConfigFiles(applicationID, environment, microserviceID, input)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	response := platform.HttpResponseConfigFilesNamesList{
		ApplicationID:  applicationID,
		Environment:    environment,
		MicroserviceID: microserviceID,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (s *service) DeleteConfigFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	applicationID := vars["applicationID"]
	microserviceID := vars["microserviceID"]
	environment := vars["environment"]

	userID := r.Header.Get("User-ID")
	customerID := r.Header.Get("Tenant-ID")
	
	logContext := s.logContext.WithFields(logrus.Fields{
		"method":          "DeleteConfigFile",
		"application_id":  applicationID,
		"microservice_id": microserviceID,
		"environment":     environment,
	})

	var input platform.HttpRequestDeleteConfigFile
	b, err := ioutil.ReadAll(r.Body)

	if err != nil {
		logContext.Info("DeleteConfigFile ERROR: Invalid request payload")
		
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(b, &input)
	if err != nil {
		logContext.Info("DeleteConfigFile BAD_REQUEST: " + err.Error())

		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, customerID, applicationID, userID)

	if !allowed {
		logContext.Info("DeleteConfigFile ERROR: not allowed")

		return
	}

	logContext.Info("Update config files")

	err = s.configFilesRepo.RemoveEntryFromConfigFiles(applicationID, environment, microserviceID, input.Key)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := platform.HttpResponseDeleteConfigFile{
		ApplicationID:  applicationID,
		Environment:    environment,
		MicroserviceID: microserviceID,
		Success:        true,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}
