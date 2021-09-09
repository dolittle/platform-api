package application

import (
	"encoding/json"
	"errors"
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

func NewService(subscriptionID string, gitRepo storage.Repo, k8sDolittleRepo platform.K8sRepo) service {
	return service{
		subscriptionID:  subscriptionID,
		gitRepo:         gitRepo,
		k8sDolittleRepo: k8sDolittleRepo,
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
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, tenantID, applicationID, userID)
	if !allowed {
		return
	}

	// This is ugly but will work "12343/"
	if !s.gitRepo.IsAutomationEnabled(tenantID, applicationID, "") {
		utils.RespondWithError(
			w,
			http.StatusBadRequest,
			fmt.Sprintf("Tenant %s with application %s does not allow changes via Studio", tenantID, applicationID),
		)
		return
	}

	application, err := s.gitRepo.GetApplication(tenantID, applicationID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Not able to find application in the storage")
		return
	}

	if err := s.validateEnvironmentDoesNotExist(input, application.Environments); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	application.Environments = append(application.Environments, input)

	err = s.gitRepo.SaveApplication(application)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to write to storage")
		return
	}

	// TODO need to create network policy for the environment
	// https://app.asana.com/0/1200181647276434/1200407495881663/f
	utils.RespondWithJSON(w, http.StatusOK, input)
}

func (s *service) validateEnvironmentDoesNotExist(inputEnvironment platform.HttpInputEnvironment, storedEnvironments []platform.HttpInputEnvironment) error {
	environmentNameExists := funk.Contains(storedEnvironments, func(storedEnvironment platform.HttpInputEnvironment) bool {
		return storedEnvironment.Name == inputEnvironment.Name
	})

	if environmentNameExists {
		return errors.New(fmt.Sprintf("Environment %s already exists", inputEnvironment.Name))
	}

	var usedDomainPrefix string
	// TODO this is not going to work with custom domains.
	// Simple logic to make sure the domainPrefix is not used
	// This is not great and should be linked to actual domains
	domainPrefixAlreadyUsed := funk.Contains(storedEnvironments, func(storedEnvironment platform.HttpInputEnvironment) bool {
		for _, storedIngress := range storedEnvironment.Ingresses {
			domainPrefixAlreadyUsed := funk.Contains(inputEnvironment.Ingresses, func(_ platform.TenantId, inputIngress platform.EnvironmentIngress) bool {
				return inputIngress.DomainPrefix == storedIngress.DomainPrefix
			})
			if domainPrefixAlreadyUsed {
				usedDomainPrefix = storedIngress.DomainPrefix
				return true
			}
		}
		return false
	})
	if domainPrefixAlreadyUsed {
		return errors.New(fmt.Sprintf("Cannot save environment %s because an ingress with domain prefix %s already exists", inputEnvironment.Name, usedDomainPrefix))
	}
	return nil
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
	// Today, we do not lookup access as we require the application rbac to be set, for now
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
	tenantInfo, err := s.gitRepo.GetTerraformTenant(tenantID)
	if err != nil {
		// TODO handle not found
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tenant := k8s.Tenant{
		ID:   tenantInfo.GUID,
		Name: tenantInfo.Name,
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

	tenantInfo, err := s.gitRepo.GetTerraformTenant(tenantID)
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

func (s *service) GetPersonalisedInfo(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("Tenant-ID")
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]

	terraformCustomer, err := s.gitRepo.GetTerraformTenant(tenantID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	terraformApplication, err := s.gitRepo.GetTerraformApplication(tenantID, applicationID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"customer":       terraformCustomer,
		"application":    terraformApplication,
		"subscriptionId": s.subscriptionID,
		"applicationId":  applicationID,
	})
}
