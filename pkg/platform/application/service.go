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
	k8sSimple "github.com/dolittle/platform-api/pkg/platform/microservice/simple/k8s"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"k8s.io/client-go/kubernetes"
)

func NewService(
	subscriptionID string,
	externalClusterHost string,
	k8sClient kubernetes.Interface,
	gitRepo storage.Repo,
	k8sDolittleRepo platformK8s.K8sRepo,
	platformOperationsImage string,
	platformEnvironment string,
	isProduction bool,
	logContext logrus.FieldLogger) service {
	return service{
		subscriptionID:          subscriptionID,
		externalClusterHost:     externalClusterHost,
		gitRepo:                 gitRepo,
		simpleRepo:              k8sSimple.NewSimpleRepo(platformEnvironment, k8sClient, k8sDolittleRepo),
		k8sDolittleRepo:         k8sDolittleRepo,
		k8sClient:               k8sClient,
		platformOperationsImage: platformOperationsImage,
		platformEnvironment:     platformEnvironment,
		isProduction:            isProduction,
		logContext:              logContext,
	}
}

func (s *service) Create(w http.ResponseWriter, r *http.Request) {
	customerID := r.Header.Get("Tenant-ID")
	logContext := s.logContext.WithFields(logrus.Fields{
		"method":      "Create",
		"customer_id": customerID,
	})

	terraformCustomer, err := s.gitRepo.GetTerraformTenant(customerID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, platform.ErrStudioInfoMissing.Error())
		return
	}

	tenant := dolittleK8s.Tenant{
		ID:   terraformCustomer.GUID,
		Name: terraformCustomer.Name,
	}

	// TODO come in via http input
	var input platform.HttpInputApplication
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

	if tenant.ID != input.TenantID {
		utils.RespondWithError(w, http.StatusBadRequest, "Tenant ID did not match")
		return
	}

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

	// TODO Confirm at least 1 environment
	// TODO we will need to revisit platform.HttpInputEnvironment
	// TODO this will overwrite
	// TODO massage the data https://app.asana.com/0/0/1201457681486811/f (sanatise the labels)

	application := storage.JSONApplication{
		ID:           input.ID,
		Name:         input.Name,
		TenantID:     tenant.ID,
		TenantName:   tenant.Name,
		Environments: make([]storage.JSONEnvironment2, 0),
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
		environmentInfo := storage.JSONEnvironment2{
			Name: environment,
			CustomerTenants: []platform.CustomerTenantInfo{
				customerTenant,
			},
			WelcomeMicroserviceID: welcomeMicroserviceID,
		}
		application.Environments = append(application.Environments, environmentInfo)
	}

	// TODO change this to use the storage
	err = s.gitRepo.SaveApplication(application)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to write to storage")
		return
	}

	platformOperationsImage := s.platformOperationsImage
	platformEnvironment := s.platformEnvironment

	resource := jobK8s.CreateApplicationResource(
		platformOperationsImage,
		platformEnvironment,
		customerID, dolittleK8s.ShortInfo{
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

func (s *service) GetLiveApplications(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("Tenant-ID")
	tenantInfo, err := s.gitRepo.GetTerraformTenant(tenantID)
	if err != nil {
		// TODO handle not found
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tenant := dolittleK8s.Tenant{
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
		Environments: funk.Map(application.Environments, func(environment storage.JSONEnvironment2) HttpResponseEnvironment {
			return HttpResponseEnvironment{
				AutomationEnabled: s.gitRepo.IsAutomationEnabledWithStudioConfig(studioInfo.StudioConfig, applicationID, environment.Name),
			}
		}).([]HttpResponseEnvironment),
		Microservices: microservices,
	}
	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (s *service) GetApplications(w http.ResponseWriter, r *http.Request) {
	customerID := r.Header.Get("Tenant-ID")

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

	response := platform.HttpResponseApplications{
		ID:           customerID,
		Name:         tenantInfo.Name,
		Applications: make([]platform.ShortInfoWithEnvironment, 0),
	}

	for _, storedApplication := range storedApplications {
		application, err := s.gitRepo.GetApplication(customerID, storedApplication.ID)
		if err != nil {
			s.logContext.WithFields(logrus.Fields{
				"error":          err,
				"method":         "s.gitRepo.GetApplication",
				"customer_id":    customerID,
				"application_id": storedApplication.ID,
			}).Error("Broken state")
			continue
		}

		for _, environmentInfo := range application.Environments {
			response.Applications = append(response.Applications, platform.ShortInfoWithEnvironment{
				ID:          storedApplication.ID,
				Name:        storedApplication.Name,
				Environment: environmentInfo.Name, // TODO do I want to include the info?
			})
		}
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (s *service) GetPersonalisedInfo(w http.ResponseWriter, r *http.Request) {
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

func (s *service) IsOnline(w http.ResponseWriter, r *http.Request) {
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
