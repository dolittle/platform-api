package purchaseorderapi

import (
	"context"
	"fmt"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	microserviceK8s "github.com/dolittle/platform-api/pkg/platform/microservice/k8s"
	"k8s.io/client-go/kubernetes"
)

type repo struct {
	k8sResource            K8sResource
	k8sClient              kubernetes.Interface
	k8sResourceSpecFactory K8sResourceSpecFactory
	k8sRepoV2              k8s.Repo
}

// NewRepo creates a new instance of purchaseorderapiRepo.
func NewRepo(k8sResource K8sResource, k8sResourceSpecFactory K8sResourceSpecFactory, k8sClient kubernetes.Interface, k8sRepoV2 k8s.Repo) Repo {
	return &repo{
		k8sResource,
		k8sClient,
		k8sResourceSpecFactory,
		k8sRepoV2,
	}
}

// Create creates a new PurchaseOrderAPI microservice
func (r *repo) Create(namespace string, customer dolittleK8s.Tenant, application dolittleK8s.Application, customerTenants []platform.CustomerTenantInfo, input platform.HttpInputPurchaseOrderInfo) error {
	// TODO not sure where this comes from really, assume dynamic

	environment := input.Environment
	microserviceID := input.Dolittle.MicroserviceID
	microserviceName := input.Name
	headImage := input.Extra.Headimage
	runtimeImage := input.Extra.Runtimeimage

	microservice := dolittleK8s.Microservice{
		ID:          microserviceID,
		Name:        microserviceName,
		Tenant:      customer,
		Application: application,
		Environment: environment,
		Kind:        platform.MicroserviceKindPurchaseOrderAPI,
	}

	ctx := context.TODO()

	if err := r.k8sResource.Create(ctx, namespace, headImage, runtimeImage, microservice, customerTenants, input.Extra); err != nil {
		return err
	}

	return nil
}

// TODO customerTenants is not in use, but we need it due to how we get the deployment name
func (r *repo) Exists(namespace string, customer dolittleK8s.Tenant, application dolittleK8s.Application, customerTenants []platform.CustomerTenantInfo, input platform.HttpInputPurchaseOrderInfo) (bool, error) {
	environment := input.Environment
	microserviceID := input.Dolittle.MicroserviceID
	microserviceName := input.Name
	headImage := input.Extra.Headimage
	runtimeImage := input.Extra.Runtimeimage

	microservice := dolittleK8s.Microservice{
		ID:          microserviceID,
		Name:        microserviceName,
		Tenant:      customer,
		Application: application,
		Environment: environment,
		Kind:        platform.MicroserviceKindPurchaseOrderAPI,
	}

	ctx := context.TODO()

	resources := r.k8sResourceSpecFactory.CreateAll(headImage, runtimeImage, microservice, customerTenants, input.Extra)
	exists, err := microserviceK8s.K8sHasDeploymentWithName(r.k8sClient, ctx, namespace, resources.Deployment.Name)
	if err != nil {
		return false, fmt.Errorf("Failed to get purchase order api deployment: %v", err)
	}
	return exists, nil
}

func (r *repo) EnvironmentHasPurchaseOrderAPI(namespace string, input platform.HttpInputPurchaseOrderInfo) (bool, error) {
	deployments, err := r.k8sRepoV2.GetDeploymentsByEnvironmentWithMicroservice(namespace, input.Environment)
	if err != nil {
		return false, err
	}

	for _, deployment := range deployments {
		if deployment.Annotations["dolittle.io/microservice-kind"] != string(platform.MicroserviceKindPurchaseOrderAPI) {
			continue
		}
		return true, nil
	}

	return false, nil
}

// Delete stops the running purchase order api and deletes the kubernetes resources.
func (r *repo) Delete(applicationID, environment, microserviceID string) error {
	ctx := context.TODO()
	return r.k8sResource.Delete(ctx, applicationID, environment, microserviceID)
}
