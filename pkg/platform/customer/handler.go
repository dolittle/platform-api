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
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/client-go/kubernetes"
)

type handler struct {
	k8sclient         kubernetes.Interface
	storageRepo       storage.RepoCustomer
	jobResourceConfig jobK8s.CreateResourceConfig
	logContext        logrus.FieldLogger
	roleBindingRepo   k8s.RepoRoleBinding
}

type HttpCustomersResponse []platform.Customer

type HttpCustomerInput struct {
	Name string `json:"name"`
}

type CustomerProvider interface {
	Create()
	GetAll()
	Get()
	Update()
}

func NewHandler(
	k8sclient kubernetes.Interface,
	storageRepo storage.RepoCustomer,
	jobResourceConfig jobK8s.CreateResourceConfig,
	logContext logrus.FieldLogger,
	roleBindingRepo k8s.RepoRoleBinding,
) handler {
	return handler{
		k8sclient:         k8sclient,
		storageRepo:       storageRepo,
		jobResourceConfig: jobResourceConfig,
		logContext:        logContext,
		roleBindingRepo:   roleBindingRepo,
	}
}

func (h *handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("User-ID")
	hasAccess, err := h.roleBindingRepo.HasUserAdminAccess(userID)
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

	logContext := h.logContext.WithFields(logrus.Fields{
		"method":      "Create",
		"customer_id": customer.ID,
	})

	err = h.storageRepo.SaveCustomer(customer)
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

	resource := jobK8s.CreateCustomerResource(h.jobResourceConfig, jobCustomer)
	err = jobK8s.DoJob(h.k8sclient, resource)
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

func (h *handler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("User-ID")
	hasAccess, err := h.roleBindingRepo.HasUserAdminAccess(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check if user has access")
		return
	}

	if !hasAccess {
		utils.RespondWithError(w, http.StatusForbidden, "You do not have access")
		return
	}

	customers, err := h.storageRepo.GetCustomers()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get customers")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, customers)
}

func IsCustomerNameValid(name string) bool {
	isValid := validation.NameIsDNSLabel(name, false)
	return len(isValid) == 0
}
