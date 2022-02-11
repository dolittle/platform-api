package customer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	jobK8s "github.com/dolittle/platform-api/pkg/platform/job/k8s"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/google/uuid"
	"github.com/thoas/go-funk"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type service struct {
	k8sclient         kubernetes.Interface
	storageRepo       storage.RepoCustomer
	jobResourceConfig jobK8s.CreateResourceConfig
}

type HttpCustomersResponse []platform.Customer

type HttpCustomerInput struct {
	Name string `json:"name"`
}

func NewService(
	k8sclient kubernetes.Interface,
	storageRepo storage.RepoCustomer,
	jobResourceConfig jobK8s.CreateResourceConfig,
) service {
	return service{
		k8sclient:         k8sclient,
		storageRepo:       storageRepo,
		jobResourceConfig: jobResourceConfig,
	}
}

func (s *service) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("User-ID")
	hasAccess := s.hasAccess(userID)

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

	err = s.storageRepo.SaveCustomer(customer)
	if err != nil {
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
	userID := r.Header.Get("User-ID")
	hasAccess := s.hasAccess(userID)

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

// hasAccess poormans security to lock down the endpoints, based on if the userID is in the rolebinding
// rolebinding is hardcoded to platform-admin
func (s *service) hasAccess(userID string) bool {
	ctx := context.TODO()
	client := s.k8sclient
	namespace := "system-api"
	roleBinding, _ := client.RbacV1().RoleBindings(namespace).Get(ctx, "platform-admin", metav1.GetOptions{})

	access := funk.Contains(roleBinding.Subjects, func(subject rbacv1.Subject) bool {
		want := rbacv1.Subject{
			Kind:     "User",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     userID,
		}
		return equality.Semantic.DeepDerivative(subject, want)
	})

	return access
}

func IsCustomerNameValid(name string) bool {
	isValid := validation.NameIsDNSLabel(name, false)
	return len(isValid) == 0
}
