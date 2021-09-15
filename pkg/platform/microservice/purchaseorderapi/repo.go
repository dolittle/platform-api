package purchaseorderapi

import (
	"context"
	"fmt"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	microserviceK8s "github.com/dolittle-entropy/platform-api/pkg/platform/microservice/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/rawdatalog"
	"k8s.io/client-go/kubernetes"
)

type repo struct {
	k8sResource            K8sResource
	rawdatalogRepo         rawdatalog.RawDataLogIngestorRepo
	k8sClient              kubernetes.Interface
	k8sResourceSpecFactory K8sResourceSpecFactory
}

// NewRepo creates a new instance of purchaseorderapiRepo.
func NewRepo(k8sResource K8sResource, k8sResourceSpecFactory K8sResourceSpecFactory, rawDataLogIngestorRepo rawdatalog.RawDataLogIngestorRepo, k8sClient kubernetes.Interface) Repo {
	return &repo{
		k8sResource,
		rawDataLogIngestorRepo,
		k8sClient,
		k8sResourceSpecFactory,
	}
}

// Create creates a new PurchaseOrderAPI microservice, and a RawDataLog and WebhookListener if they don't exist.
func (r *repo) Create(namespace string, tenant k8s.Tenant, application k8s.Application, input platform.HttpInputPurchaseOrderInfo) error {
	// TODO not sure where this comes from really, assume dynamic

	environment := input.Environment
	microserviceID := input.Dolittle.MicroserviceID
	microserviceName := input.Name
	headImage := input.Extra.Headimage
	runtimeImage := input.Extra.Runtimeimage

	microservice := k8s.Microservice{
		ID:          microserviceID,
		Name:        microserviceName,
		Tenant:      tenant,
		Application: application,
		Environment: environment,
		ResourceID:  microserviceK8s.TodoCustomersTenantID,
		Kind:        platform.MicroserviceKindPurchaseOrderAPI,
	}

	ctx := context.TODO()

	// if err := r.rawdatalogRepo.EnsureForPurchaseOrderAPI(namespace, environment, tenant, application, input.Extra.Webhooks); err != nil {
	// 	return err
	// }

	if err := r.k8sResource.Create(namespace, headImage, runtimeImage, microservice, input.Extra, ctx); err != nil {
		return err
	}

	return nil
}
func (r *repo) Exists(namespace string, tenant k8s.Tenant, application k8s.Application, input platform.HttpInputPurchaseOrderInfo) (bool, error) {
	// TODO not sure where this comes from really, assume dynamic

	environment := input.Environment
	microserviceID := input.Dolittle.MicroserviceID
	microserviceName := input.Name
	headImage := input.Extra.Headimage
	runtimeImage := input.Extra.Runtimeimage

	microservice := k8s.Microservice{
		ID:          microserviceID,
		Name:        microserviceName,
		Tenant:      tenant,
		Application: application,
		Environment: environment,
		ResourceID:  microserviceK8s.TodoCustomersTenantID,
		Kind:        platform.MicroserviceKindPurchaseOrderAPI,
	}

	ctx := context.TODO()

	deployment, err := microserviceK8s.K8sGetDeployment(r.k8sClient, ctx, namespace, microserviceID)
	if err != nil {
		return false, fmt.Errorf("Failed to get purchase order api deployment: %v", err)
	}
	resources := r.k8sResourceSpecFactory.CreateAll(headImage, runtimeImage, microservice, input.Extra)
	return resources.Deployment.Name == deployment.Name, nil
}

// Delete stops the running purchase order api and deletes the kubernetes resources.
func (r *repo) Delete(namespace string, microserviceID string) error {
	ctx := context.TODO()
	return r.k8sResource.Delete(namespace, microserviceID, ctx)
}
