package businessmoment

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/businessmomentsadaptor"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

func NewService(logContext logrus.FieldLogger, gitRepo storage.Repo, k8sDolittleRepo platform.K8sRepo, k8sClient *kubernetes.Clientset) service {
	return service{
		logContext:            logContext,
		gitRepo:               gitRepo,
		k8sDolittleRepo:       k8sDolittleRepo,
		k8sClient:             k8sClient,
		k8sBusinessMomentRepo: businessmomentsadaptor.NewK8sRepo(k8sClient),
	}
}

func (s *service) DeleteMoment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	environment := strings.ToLower(vars["environment"])
	microserviceID := vars["microserviceID"]
	momentID := strings.ToLower(vars["momentID"])

	logContext := s.logContext.WithFields(logrus.Fields{
		"application_id":  applicationID,
		"environment":     environment,
		"microservice_id": microserviceID,
		"moment_id":       momentID,
	})

	userID := r.Header.Get("User-ID")
	tenantID := r.Header.Get("Tenant-ID")
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, tenantID, applicationID, userID)
	if !allowed {
		return
	}

	err := s.gitRepo.DeleteBusinessMoment(tenantID, applicationID, environment, microserviceID, momentID)
	if err != nil {
		// TODO handle if error not found?
		logContext.WithFields(logrus.Fields{
			"error":  err,
			"method": "s.gitRepo.DeleteBusinessMoment",
		}).Error("request")
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong")
		return
	}

	err = s.eventUpdateConfigmap(tenantID, applicationID, environment, microserviceID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong whilst updating business moments to microservice")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message":        "Moment removed",
		"tenant_id":      tenantID,
		"application_id": applicationID,
		"environment":    environment,
		"moment_id":      momentID,
		"action":         "Remove business moment",
	})
}

func (s *service) DeleteEntity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	environment := strings.ToLower(vars["environment"])
	microserviceID := vars["microserviceID"]
	entityID := strings.ToLower(vars["entityID"])

	logContext := s.logContext.WithFields(logrus.Fields{
		"application_id":  applicationID,
		"environment":     environment,
		"microservice_id": microserviceID,
		"entity_id":       entityID,
	})

	userID := r.Header.Get("User-ID")
	tenantID := r.Header.Get("Tenant-ID")
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, tenantID, applicationID, userID)
	if !allowed {
		return
	}

	err := s.gitRepo.DeleteBusinessMomentEntity(tenantID, applicationID, environment, microserviceID, entityID)
	if err != nil {
		// TODO handle if error not found?
		logContext.WithFields(logrus.Fields{
			"error":  err,
			"method": "s.gitRepo.DeleteBusinessMomentEntity",
		}).Error("request")
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong")
		return
	}

	err = s.eventUpdateConfigmap(tenantID, applicationID, environment, microserviceID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong whilst updating business moments to microservice")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message":        "Entity removed",
		"tenant_id":      tenantID,
		"application_id": applicationID,
		"environment":    environment,
		"entity_id":      entityID,
		"action":         "Remove business moment entity and its moments",
	})
}

func (s *service) SaveEntity(w http.ResponseWriter, r *http.Request) {
	var input platform.HttpInputBusinessMomentEntity
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(b, &input)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	userID := r.Header.Get("User-ID")
	tenantID := r.Header.Get("Tenant-ID")
	applicationID := input.ApplicationID
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, tenantID, applicationID, userID)
	if !allowed {
		return
	}

	rawBytes, err := s.gitRepo.GetMicroservice(tenantID, applicationID, input.Environment, input.MicroserviceID)
	if err != nil {
		// TODO add logContext
		utils.RespondWithError(w, http.StatusBadRequest, "Not able to find microservice in the storage")
		return
	}

	var microservice platform.HttpMicroserviceBase
	err = json.Unmarshal(rawBytes, &microservice)

	if err != nil {
		// TODO add logContext
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong")
		return
	}

	if microservice.Kind != platform.MicroserviceKindBusinessMomentsAdaptor {
		utils.RespondWithError(w, http.StatusBadRequest, "Not Business moment to find microservice in the storage")
		return
	}

	err = s.gitRepo.SaveBusinessMomentEntity(tenantID, input)
	if err != nil {
		// TODO add logContext
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong")
		return
	}

	err = s.eventUpdateConfigmap(tenantID, input.ApplicationID, input.Environment, input.MicroserviceID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong whilst updating business moments to microservice")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, input)
}

func (s *service) SaveMoment(w http.ResponseWriter, r *http.Request) {
	var input platform.HttpInputBusinessMoment
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(b, &input)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	userID := r.Header.Get("User-ID")
	tenantID := r.Header.Get("Tenant-ID")
	applicationID := input.ApplicationID
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, tenantID, applicationID, userID)
	if !allowed {
		return
	}

	rawBytes, err := s.gitRepo.GetMicroservice(tenantID, applicationID, input.Environment, input.MicroserviceID)
	if err != nil {
		// TODO add logContext
		utils.RespondWithError(w, http.StatusBadRequest, "Not able to find microservice in the storage")
		return
	}

	var microservice platform.HttpMicroserviceBase
	err = json.Unmarshal(rawBytes, &microservice)

	if err != nil {
		// TODO add logContext
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong")
		return
	}

	if microservice.Kind != platform.MicroserviceKindBusinessMomentsAdaptor {
		utils.RespondWithError(w, http.StatusBadRequest, "Not Business moment to find microservice in the storage")
		return
	}

	err = s.gitRepo.SaveBusinessMoment(tenantID, input)
	if err != nil {
		// TODO add logContext
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong")
		return
	}

	err = s.eventUpdateConfigmap(tenantID, applicationID, input.Environment, input.MicroserviceID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong whilst updating business moments to microservice")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, input)
}

func (s *service) GetMoments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	environment := strings.ToLower(vars["environment"])

	userID := r.Header.Get("User-ID")
	tenantID := r.Header.Get("Tenant-ID")
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, tenantID, applicationID, userID)
	if !allowed {
		return
	}

	data, err := s.gitRepo.GetBusinessMoments(tenantID, applicationID, environment)
	if err != nil {
		// TODO add logContext
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, data)
}

func (s *service) eventUpdateConfigmap(tenantID string, applicationID string, environment string, microserviceID string) error {
	logContext := s.logContext
	environment = strings.ToLower(environment)

	//  TODO this should be an event
	data, err := s.gitRepo.GetBusinessMoments(tenantID, applicationID, environment)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error":  err,
			"method": "s.gitRepo.GetBusinessMoments",
		}).Error("issue updating configmap")
		return err
	}

	configMap, err := s.k8sBusinessMomentRepo.GetBusinessMomentsConfigmap(applicationID, environment, microserviceID)
	if err != nil {
		// TODO defend and make?
		logContext.WithFields(logrus.Fields{
			"error":  err,
			"method": "s.k8sBusinessMomentRepo.GetBusinessMomentsConfigmap",
		}).Error("issue updating configmap")
		return err
	}

	dataBytes, _ := json.Marshal(data)
	err = s.k8sBusinessMomentRepo.SaveBusinessMomentsConfigmap(configMap, dataBytes)
	if err != nil {
		// TODO
		logContext.WithFields(logrus.Fields{
			"error":  err,
			"method": "s.k8sBusinessMomentRepo.SaveBusinessMomentsConfigmap",
		}).Error("issue updating configmap")
		return err
	}

	return nil
}
