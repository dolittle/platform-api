package customer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	jobK8s "github.com/dolittle/platform-api/pkg/platform/job/k8s"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/google/uuid"
	"k8s.io/client-go/kubernetes"
)

func NewService(
	k8sclient kubernetes.Interface,
	storageRepo storage.RepoCustomer,
	platformOperationsImage string,
	platformEnvironment string,
) service {
	return service{
		k8sclient:               k8sclient,
		storageRepo:             storageRepo,
		platformOperationsImage: platformOperationsImage,
		platformEnvironment:     platformEnvironment,
	}
}

func (s *service) Create(w http.ResponseWriter, r *http.Request) {
	// TODO need to figure out how we might want to expose this
	disabled := true
	if disabled {
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

	customer := storage.JSONCustomer{
		ID:   uuid.New().String(),
		Name: input.Name,
	}

	err = s.storageRepo.SaveCustomer(customer)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save customer")
		return
	}

	platformOperationsImage := s.platformOperationsImage
	platformEnvironment := s.platformEnvironment
	jobCustomer := dolittleK8s.ShortInfo{
		ID:   customer.ID,
		Name: customer.Name,
	}

	resource := jobK8s.CreateCustomerResource(platformOperationsImage, platformEnvironment, jobCustomer)
	err = jobK8s.DoJob(s.k8sclient, resource)
	if err != nil {
		// TODO log that we failed to make the job
		fmt.Println(err)
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
	// TODO need to figure out how we might want to expose this
	disabled := true
	if disabled {
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
