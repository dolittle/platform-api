package application

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/thoas/go-funk"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

func NewService(gitRepo storage.Repo, k8sDolittleRepo platform.K8sRepo) service {
	return service{
		gitRepo:         gitRepo,
		k8sDolittleRepo: k8sDolittleRepo,
		//k8sClient:       k8sClient,
	}
}

func (s *service) SaveEnvironment(w http.ResponseWriter, r *http.Request) {
	var input platform.HttpInputEnvironment
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
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

	applicationID := input.ApplicationID
	applicationInfo, err := s.k8sDolittleRepo.GetApplication(applicationID)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			fmt.Println(err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		utils.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Application %s not found", applicationID))
		return
	}

	tenantID := applicationInfo.Tenant.ID
	if tenantID != "453e04a7-4f9d-42f2-b36c-d51fa2c83fa3" {
		utils.RespondWithError(w, http.StatusBadRequest, "Currently locked down to tenant 453e04a7-4f9d-42f2-b36c-d51fa2c83fa3")
		return
	}

	allowed, err := s.k8sDolittleRepo.CanModifyApplication(tenantID, applicationInfo.ID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !allowed {
		utils.RespondWithError(w, http.StatusForbidden, "You are not allowed to make this request")
		return
	}

	application, err := s.gitRepo.GetApplication(tenantID, applicationID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Not able to find application in the storage")
		return
	}

	// TODO this is not going to work with custom domains.
	// Simple logic to make sure the domainPrefix is not used
	// This is not great and should be linked to actual domains
	exists := funk.Contains(application.Environments, func(environment platform.HttpInputEnvironment) bool {
		found := false
		if environment.Name == input.Name {
			found = true
		}
		if environment.DomainPrefix == input.DomainPrefix {
			found = true
		}
		return found
	})

	if exists {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Environment %s already exists", input.Name))
		return
	}

	application.Environments = append(application.Environments, input)

	err = s.gitRepo.SaveApplication(application)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to write to storage")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, input)
}

func (s *service) Create(w http.ResponseWriter, r *http.Request) {
	var input platform.HttpInputApplication
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(b, &input)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	_, err = s.gitRepo.GetApplication(input.TenantID, input.ID)
	if err == nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Application already exists")
		return
	}

	// TODO this will overwrite
	application := platform.HttpResponseApplication{
		ID:           input.ID,
		Name:         input.Name,
		TenantID:     input.TenantID,
		Environments: make([]platform.HttpInputEnvironment, 0),
	}

	err = s.gitRepo.SaveApplication(application)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to write to storage")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, input)
}

func (s *service) GetLiveApplications(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("Tenant-ID")

	// TODO get tenant from syncing the terraform output into the repo (which we might have access to if we use the same repo)
	tenantInfo, err := s.gitRepo.GetTenant(tenantID)
	if err != nil {
		// TODO handle not found
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tenant := k8s.Tenant{
		ID:   tenantInfo.GUID,
		Name: tenantInfo.Name,
	}

	if tenant.ID != "453e04a7-4f9d-42f2-b36c-d51fa2c83fa3" {
		utils.RespondWithError(w, http.StatusBadRequest, "Currently locked down to tenant 453e04a7-4f9d-42f2-b36c-d51fa2c83fa3")
		return
	}

	liveApplications, err := s.k8sDolittleRepo.GetApplications(tenantID)
	if err != nil {
		// TODO change
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Lookup environments
	response := platform.HttpResponseApplications{
		ID:   tenantID,
		Name: tenant.Name,
	}

	for _, liveApplication := range liveApplications {
		application, err := s.gitRepo.GetApplication(tenantID, liveApplication.ID)
		if err != nil {
			// TODO change
			utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		for _, environmentInfo := range application.Environments {
			response.Applications = append(response.Applications, platform.ShortInfoWithEnvironment{
				ID:          liveApplication.ID,
				Name:        liveApplication.Name,
				Environment: environmentInfo.Name, // TODO do I want to include the info?
			})
		}
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (s *service) GetByID(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("Tenant-ID")
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]

	application, err := s.gitRepo.GetApplication(tenantID, applicationID)
	if err != nil {
		// TODO check if not found
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	microservices, err := s.gitRepo.GetMicroservices(tenantID, applicationID)
	utils.RespondWithJSON(w, http.StatusOK, platform.HttpResponseApplication2{
		ID:            application.ID,
		Name:          application.Name,
		TenantID:      application.TenantID,
		Environments:  application.Environments,
		Microservices: microservices,
	})
}

func (s *service) GetApplications(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("Tenant-ID")

	tenantInfo, err := s.gitRepo.GetTenant(tenantID)
	if err != nil {
		// TODO handle not found
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	storedApplications, err := s.gitRepo.GetApplications(tenantID)
	if err != nil {
		// TODO check if not found
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := platform.HttpResponseApplications{
		ID:   tenantID,
		Name: tenantInfo.Name,
	}

	for _, storedApplication := range storedApplications {
		application, err := s.gitRepo.GetApplication(tenantID, storedApplication.ID)
		if err != nil {
			// TODO change
			utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		for _, environmentInfo := range application.Environments {
			response.Applications = append(response.Applications, platform.ShortInfoWithEnvironment{
				ID:          storedApplication.ID,
				Name:        storedApplication.Name,
				Environment: environmentInfo.Name, // TODO do I want to include the info?
			})
		}
	}

	//microservices, err := s.gitRepo.GetMicroservices(tenantID, applicationID)
	utils.RespondWithJSON(w, http.StatusOK, response)
}
