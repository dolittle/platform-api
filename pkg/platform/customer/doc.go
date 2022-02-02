package customer

import (
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"k8s.io/client-go/kubernetes"
)

type service struct {
	k8sclient               kubernetes.Interface
	storageRepo             storage.RepoCustomer
	platformOperationsImage string
	platformEnvironment     string
}

type HttpCustomersResponse []platform.Customer

type HttpCustomerInput struct {
	Name string `json:"name"`
}
