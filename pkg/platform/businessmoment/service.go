package businessmoment

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/gorilla/mux"
)

func NewService(gitRepo storage.Repo, k8sDolittleRepo platform.K8sRepo) service {
	return service{
		gitRepo:         gitRepo,
		k8sDolittleRepo: k8sDolittleRepo}
}

func (s *service) DeleteMoment(w http.ResponseWriter, r *http.Request) {
	// Maybe make this include a body
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	environment := strings.ToLower(vars["environment"])
	microserviceID := vars["microserviceID"]
	momentID := strings.ToLower(vars["momentID"])

	// TODO add checks
	userID := r.Header.Get("User-ID")
	if userID == "" {
		// If the middleware is enabled this shouldn't happen
		utils.RespondWithError(w, http.StatusForbidden, "User-ID is missing from the headers")
		return
	}

	// This doesnt care for kubernetes yet
	tenantID := r.Header.Get("Tenant-ID")

	// TODO could be helper function
	allowed, err := s.k8sDolittleRepo.CanModifyApplication(tenantID, applicationID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !allowed {
		utils.RespondWithError(w, http.StatusForbidden, "You are not allowed to make this request")
		return
	}
	//

	err = s.gitRepo.DeleteBusinessMoment(tenantID, applicationID, environment, microserviceID, momentID)
	if err != nil {
		// TODO add logContext
		// TODO handle if error not found?
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong")
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
	// Maybe make this include a body
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	environment := strings.ToLower(vars["environment"])
	microserviceID := vars["microserviceID"]
	entityID := strings.ToLower(vars["entityID"])

	// TODO add checks
	userID := r.Header.Get("User-ID")
	if userID == "" {
		// If the middleware is enabled this shouldn't happen
		utils.RespondWithError(w, http.StatusForbidden, "User-ID is missing from the headers")
		return
	}

	// This doesnt care for kubernetes yet
	tenantID := r.Header.Get("Tenant-ID")

	// TODO could be helper function
	allowed, err := s.k8sDolittleRepo.CanModifyApplication(tenantID, applicationID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !allowed {
		utils.RespondWithError(w, http.StatusForbidden, "You are not allowed to make this request")
		return
	}
	//

	err = s.gitRepo.DeleteBusinessMomentEntity(tenantID, applicationID, environment, microserviceID, entityID)
	if err != nil {
		// TODO add logContext
		// TODO handle if error not found?
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong")
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
	if userID == "" {
		// If the middleware is enabled this shouldn't happen
		utils.RespondWithError(w, http.StatusForbidden, "User-ID is missing from the headers")
		return
	}

	// This doesnt care for kubernetes yet
	tenantID := r.Header.Get("Tenant-ID")
	applicationID := input.ApplicationID

	// TODO could be helper function
	allowed, err := s.k8sDolittleRepo.CanModifyApplication(tenantID, applicationID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !allowed {
		utils.RespondWithError(w, http.StatusForbidden, "You are not allowed to make this request")
		return
	}
	//

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

	if microservice.Kind != platform.BusinessMomentsAdaptor {
		utils.RespondWithError(w, http.StatusBadRequest, "Not Business moment to find microservice in the storage")
		return
	}

	err = s.gitRepo.SaveBusinessMomentEntity(tenantID, input)
	if err != nil {
		// TODO add logContext
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong")
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
	if userID == "" {
		// If the middleware is enabled this shouldn't happen
		utils.RespondWithError(w, http.StatusForbidden, "User-ID is missing from the headers")
		return
	}

	// This doesnt care for kubernetes yet
	tenantID := r.Header.Get("Tenant-ID")
	applicationID := input.ApplicationID

	fmt.Println(applicationID, userID)
	// TODO could be helper function
	allowed, err := s.k8sDolittleRepo.CanModifyApplication(tenantID, applicationID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !allowed {
		utils.RespondWithError(w, http.StatusForbidden, "You are not allowed to make this request")
		return
	}
	//

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

	if microservice.Kind != platform.BusinessMomentsAdaptor {
		utils.RespondWithError(w, http.StatusBadRequest, "Not Business moment to find microservice in the storage")
		return
	}

	err = s.gitRepo.SaveBusinessMoment(tenantID, input)
	if err != nil {
		// TODO add logContext
		utils.RespondWithError(w, http.StatusInternalServerError, "Something has gone wrong")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, input)
}

func (s *service) GetMoments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	environment := strings.ToLower(vars["environment"])

	userID := r.Header.Get("User-ID")
	if userID == "" {
		// If the middleware is enabled this shouldn't happen
		utils.RespondWithError(w, http.StatusForbidden, "User-ID is missing from the headers")
		return
	}

	tenantID := r.Header.Get("Tenant-ID")

	// TODO could be helper function
	allowed, err := s.k8sDolittleRepo.CanModifyApplication(tenantID, applicationID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !allowed {
		utils.RespondWithError(w, http.StatusForbidden, "You are not allowed to make this request")
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
