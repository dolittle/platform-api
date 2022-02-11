package application

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	jobK8s "github.com/dolittle/platform-api/pkg/platform/job/k8s"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/simple"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/client-go/kubernetes"
)

type Service struct {
	subscriptionID      string
	externalClusterHost string
	simpleRepo          simple.Repo
	gitRepo             storage.Repo
	k8sDolittleRepo     platformK8s.K8sRepo
	k8sClient           kubernetes.Interface
	logContext          logrus.FieldLogger
	jobResourceConfig   jobK8s.CreateResourceConfig
}

func NewService(
	subscriptionID string,
	externalClusterHost string,
	k8sClient kubernetes.Interface,
	gitRepo storage.Repo,
	k8sDolittleRepo platformK8s.K8sRepo,
	jobResourceConfig jobK8s.CreateResourceConfig,
	simpleRepo simple.Repo,
	logContext logrus.FieldLogger) Service {
	return Service{
		subscriptionID:      subscriptionID,
		externalClusterHost: externalClusterHost,
		gitRepo:             gitRepo,
		jobResourceConfig:   jobResourceConfig,
		simpleRepo:          simpleRepo,
		k8sDolittleRepo:     k8sDolittleRepo,
		k8sClient:           k8sClient,
		logContext:          logContext,
	}
}

func (s *Service) Create(w http.ResponseWriter, r *http.Request) {
	customerID := r.Header.Get("Tenant-ID")
	logContext := s.logContext.WithFields(logrus.Fields{
		"method":      "Create",
		"customer_id": customerID,
	})

	studioConfig, err := s.gitRepo.GetStudioConfig(customerID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !studioConfig.CanCreateApplication {
		utils.RespondWithError(w, http.StatusForbidden, "Creating applications is disabled")
		return
	}

	terraformCustomer, err := s.gitRepo.GetTerraformTenant(customerID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, platform.ErrStudioInfoMissing.Error())
		return
	}

	customer := dolittleK8s.Tenant{
		ID:   terraformCustomer.GUID,
		Name: terraformCustomer.Name,
	}

	var input HttpInputApplication
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Bad input")
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(b, &input)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Confirm at least 1 environment
	if len(input.Environments) == 0 {
		utils.RespondWithError(w, http.StatusUnprocessableEntity, "You need at least one environment")
		return
	}

	if !IsApplicationNameValid(input.Name) {
		utils.RespondWithError(w, http.StatusUnprocessableEntity, "Application name is not valid")
		return
	}

	// Confirm application id not already in use for this customer
	current, err := s.gitRepo.GetApplication(customerID, input.ID)
	if err != nil {
		if err != storage.ErrNotFound {
			logContext.WithFields(logrus.Fields{
				"error":          err,
				"application_id": input.ID,
			}).Error("Storage has failed")
			utils.RespondWithError(w, http.StatusInternalServerError, platform.ErrStudioInfoMissing.Error())
			return
		}
	}

	if current.ID == input.ID {
		utils.RespondWithError(w, http.StatusBadRequest, "Application id already exists")
		return
	}
	// TODO do we need to confirm applicationId is unique in the cluster?

	// TODO this will overwrite
	// TODO massage the data https://app.asana.com/0/0/1201457681486811/f (sanatise the labels)

	application := storage.JSONApplication{
		ID:           input.ID,
		Name:         input.Name,
		TenantID:     customer.ID,
		TenantName:   customer.Name,
		Environments: make([]storage.JSONEnvironment, 0),
		Status: storage.JSONBuildStatus{
			State:     storage.BuildStatusStateWaiting,
			StartedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}

	environments := input.Environments

	for _, environment := range environments {
		// TODO this could development microserviceID const (ask @joel)
		welcomeMicroserviceID := uuid.New().String()
		customerTenant := dolittleK8s.NewDevelopmentCustomerTenantInfo(environment, welcomeMicroserviceID)
		environmentInfo := storage.JSONEnvironment{
			Name: environment,
			CustomerTenants: []platform.CustomerTenantInfo{
				customerTenant,
			},
			WelcomeMicroserviceID: welcomeMicroserviceID,
		}
		application.Environments = append(application.Environments, environmentInfo)
	}

	err = s.gitRepo.SaveApplication(application)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to write to storage")
		return
	}

	resource := jobK8s.CreateApplicationResource(
		s.jobResourceConfig,
		customerID,
		dolittleK8s.ShortInfo{
			ID:   application.ID,
			Name: application.Name,
		})

	err = jobK8s.DoJob(s.k8sClient, resource)
	if err != nil {
		// TODO log that we failed to make the job
		logContext.WithFields(logrus.Fields{
			"error":          err,
			"application_id": application.ID,
		}).Error("Failed to create job to create application")
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create application")
		return
	}

	// TODO Do I need to send the jobId as it is made up of the applicationID?
	utils.RespondWithJSON(
		w,
		http.StatusOK,
		map[string]string{
			"jobId": resource.ObjectMeta.Name,
		},
	)
}

func (s *Service) GetLiveApplications(w http.ResponseWriter, r *http.Request) {
	customerID := r.Header.Get("Tenant-ID")
	studioConfig, err := s.gitRepo.GetStudioConfig(customerID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tenantInfo, err := s.gitRepo.GetTerraformTenant(customerID)
	if err != nil {
		// TODO handle not found
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	customer := dolittleK8s.Tenant{
		ID:   tenantInfo.GUID,
		Name: tenantInfo.Name,
	}

	liveApplications, err := s.k8sDolittleRepo.GetApplications(customerID)
	if err != nil {
		// TODO change
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Lookup environments
	response := HttpResponseApplications{
		ID:                   customer.ID,
		Name:                 customer.Name,
		CanCreateApplication: studioConfig.CanCreateApplication,
	}

	for _, liveApplication := range liveApplications {
		application, err := s.gitRepo.GetApplication(customerID, liveApplication.ID)
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

func (s *Service) GetByID(w http.ResponseWriter, r *http.Request) {
	logContext := s.logContext.WithFields(logrus.Fields{
		"method": "GetByID",
	})
	customerID := r.Header.Get("Tenant-ID")
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]

	studioInfo, err := storage.GetStudioInfo(s.gitRepo, customerID, applicationID, logContext)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	application, err := s.gitRepo.GetApplication(customerID, applicationID)
	if err != nil {
		// TODO check if not found
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	microservices, err := s.gitRepo.GetMicroservices(customerID, applicationID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := HttpResponseApplication{
		Name:       application.Name,
		ID:         application.ID,
		TenantID:   studioInfo.TerraformCustomer.GUID,
		TenantName: studioInfo.TerraformCustomer.Name,
		Environments: funk.Map(application.Environments, func(environment storage.JSONEnvironment) HttpResponseEnvironment {
			return HttpResponseEnvironment{
				AutomationEnabled: s.gitRepo.IsAutomationEnabledWithStudioConfig(studioInfo.StudioConfig, applicationID, environment.Name),
				Name:              environment.Name,
			}
		}).([]HttpResponseEnvironment),
		Microservices: microservices,
	}
	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (s *Service) GetApplications(w http.ResponseWriter, r *http.Request) {
	customerID := r.Header.Get("Tenant-ID")

	studioConfig, err := s.gitRepo.GetStudioConfig(customerID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tenantInfo, err := s.gitRepo.GetTerraformTenant(customerID)
	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error":       err,
			"method":      "s.gitRepo.GetTerraformTenant",
			"customer_id": customerID,
		}).Error("Broken state")
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	storedApplications, err := s.gitRepo.GetApplications(customerID)
	if err != nil {
		if err != storage.ErrNotFound {
			utils.RespondWithError(w, http.StatusNotFound, "No applications")
		}
		s.logContext.WithFields(logrus.Fields{
			"error":       err,
			"method":      "s.gitRepo.GetApplications",
			"customer_id": customerID,
		}).Error("Broken state")
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := HttpResponseApplications{
		ID:                   customerID,
		Name:                 tenantInfo.Name,
		CanCreateApplication: studioConfig.CanCreateApplication,
		Applications:         make([]platform.ShortInfoWithEnvironment, 0),
	}

	for _, storedApplication := range storedApplications {
		for _, environmentInfo := range storedApplication.Environments {
			response.Applications = append(response.Applications, platform.ShortInfoWithEnvironment{
				ID:          storedApplication.ID,
				Name:        storedApplication.Name,
				Environment: environmentInfo.Name,
			})
		}
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (s *Service) GetPersonalisedInfo(w http.ResponseWriter, r *http.Request) {
	logContext := s.logContext.WithFields(logrus.Fields{
		"method": "GetPersonalisedInfo",
	})
	customerID := r.Header.Get("Tenant-ID")
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]

	studioInfo, err := storage.GetStudioInfo(s.gitRepo, customerID, applicationID, logContext)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	clusterEndpoint := s.externalClusterHost
	// TODO https://app.asana.com/0/1201325052247030/1201756581211961/f
	resourceGroup := "Infrastructure-Essential"
	clusterName := "Cluster-Production-Three"
	utils.RespondWithJSON(w, http.StatusOK, platform.HttpResponsePersonalisedInfo{
		ResourceGroup:         resourceGroup,
		ClusterName:           clusterName,
		SubscriptionID:        s.subscriptionID,
		ApplicationID:         applicationID,
		ContainerRegistryName: studioInfo.TerraformCustomer.ContainerRegistryName,
		Endpoints: platform.HttpResponsePersonalisedInfoEndpoints{
			Cluster:           clusterEndpoint,
			ContainerRegistry: fmt.Sprintf("%s.azurecr.io", studioInfo.TerraformCustomer.ContainerRegistryName),
		},
	})
}

func (s *Service) IsOnline(w http.ResponseWriter, r *http.Request) {
	customerID := r.Header.Get("Tenant-ID")
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]

	application, err := s.gitRepo.GetApplication(customerID, applicationID)
	if err != nil {
		if err == storage.ErrNotFound {
			utils.RespondWithError(w, http.StatusNotFound, "Application id does not exist in our platform")
			return
		}
		// TODO check if not found
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// TODO consider not using storage as the response
	utils.RespondWithJSON(w, http.StatusOK, application.Status)
}

func IsApplicationNameValid(name string) bool {
	isValid := validation.NameIsDNSLabel(name, false)
	return len(isValid) == 0
}
