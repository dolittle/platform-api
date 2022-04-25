package customer

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	jobK8s "github.com/dolittle/platform-api/pkg/platform/job/k8s"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/client-go/kubernetes"
)

type HTTPResponseCustomer struct {
	Customer     platform.Customer                   `json:"customer"`
	Applications []platform.ShortInfoWithEnvironment `json:"applications"`
	StudioConfig HTTPStudioConfig                    `json:"studioConfig"`
}

type HTTPStudioConfig struct {
	BuildOverwrite       bool     `json:"buildOverwrite"`
	DisabledEnvironments []string `json:"disabledEnvironments"`
	CanCreateApplication bool     `json:"canCreateApplication"`
}

type CustomerRepo interface {
	GetCustomers() ([]platform.Customer, error)
	SaveCustomer(customer storage.JSONCustomer) error
	GetApplications(customerID string) ([]storage.JSONApplication, error)
	GetStudioConfig(customerID string) (platform.StudioConfig, error)
}

type service struct {
	k8sclient         kubernetes.Interface
	storageRepo       CustomerRepo
	jobResourceConfig jobK8s.CreateResourceConfig
	logContext        logrus.FieldLogger
	roleBindingRepo   k8s.RepoRoleBinding
}

type HttpCustomersResponse []platform.Customer

type HttpCustomerInput struct {
	Name string `json:"name"`
}

func NewService(
	k8sclient kubernetes.Interface,
	storageRepo CustomerRepo,
	jobResourceConfig jobK8s.CreateResourceConfig,
	logContext logrus.FieldLogger,
	roleBindingRepo k8s.RepoRoleBinding,
) service {
	return service{
		k8sclient:         k8sclient,
		storageRepo:       storageRepo,
		jobResourceConfig: jobResourceConfig,
		logContext:        logContext,
		roleBindingRepo:   roleBindingRepo,
	}
}

func (s *service) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("User-ID")
	hasAccess, err := s.roleBindingRepo.HasUserAdminAccess(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check if user has access")
		return
	}

	if !hasAccess {
		utils.RespondWithError(w, http.StatusForbidden, "You do not have access")
		return
	}

	var input HttpCustomerInput
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err = json.Unmarshal(b, &input)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if !IsCustomerNameValid(input.Name) {
		utils.RespondWithError(w, http.StatusUnprocessableEntity, "Customer name is not valid")
		return
	}

	customer := storage.JSONCustomer{
		ID:   uuid.New().String(),
		Name: input.Name,
	}

	logContext := s.logContext.WithFields(logrus.Fields{
		"method":      "Create",
		"customer_id": customer.ID,
	})

	err = s.storageRepo.SaveCustomer(customer)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to save customer")
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save customer")
		return
	}

	jobCustomer := dolittleK8s.ShortInfo{
		ID:   customer.ID,
		Name: customer.Name,
	}

	resource := jobK8s.CreateCustomerResource(s.jobResourceConfig, jobCustomer)
	err = jobK8s.DoJob(s.k8sclient, resource)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to create job to create application")
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save customer")
		return
	}

	utils.RespondWithJSON(
		w,
		http.StatusOK,
		map[string]string{
			"jobId": resource.ObjectMeta.Name,
		},
	)
}

func (s *service) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("User-ID")
	hasAccess, err := s.roleBindingRepo.HasUserAdminAccess(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check if user has access")
		return
	}

	if !hasAccess {
		utils.RespondWithError(w, http.StatusForbidden, "You do not have access")
		return
	}

	customers, err := s.storageRepo.GetCustomers()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get customers")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, customers)
}

func (s *service) GetOne(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("User-ID")
	hasAccess, err := s.roleBindingRepo.HasUserAdminAccess(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check if user has access")
		return
	}

	if !hasAccess {
		utils.RespondWithError(w, http.StatusForbidden, "You do not have access")
		return
	}

	vars := mux.Vars(r)
	customerID := vars["customerID"]

	// TODO https://app.asana.com/0/0/1202088245934608/f
	studioConfig, err := s.storageRepo.GetStudioConfig(customerID)

	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	customers, err := s.storageRepo.GetCustomers()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get customers")
		return
	}

	found := funk.Find(customers, func(customer platform.Customer) bool {
		return customer.ID == customerID
	})

	if found == nil {
		utils.RespondWithError(w, http.StatusNotFound, "Failed to get customer")
		return
	}

	storedApplications, err := s.storageRepo.GetApplications(customerID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get customers")
		return
	}

	response := HTTPResponseCustomer{
		Customer:     found.(platform.Customer),
		Applications: make([]platform.ShortInfoWithEnvironment, 0),
		StudioConfig: HTTPStudioConfig{
			BuildOverwrite:       studioConfig.BuildOverwrite,
			DisabledEnvironments: studioConfig.DisabledEnvironments,
			CanCreateApplication: studioConfig.CanCreateApplication,
		},
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

func IsCustomerNameValid(name string) bool {
	isValid := validation.NameIsDNSLabel(name, false)
	return len(isValid) == 0
}
