package purchaseorderapi

import (
	"context"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/rawdatalog"
)

type repo struct {
	k8sResource    K8sResource
	rawdatalogRepo rawdatalog.RawDataLogIngestorRepo
}

// NewRepo creates a new instance of purchaseorderapiRepo.
func NewRepo(k8sResource K8sResource, rawDataLogIngestorRepo rawdatalog.RawDataLogIngestorRepo) Repo {
	return &repo{
		k8sResource,
		rawDataLogIngestorRepo,
	}
}

// Create creates a new PurchaseOrderAPI microservice, and a RawDataLog and WebhookListener if they don't exist.
func (r *repo) Create(namespace string, customer k8s.Tenant, application k8s.Application, tenant platform.TenantId, input platform.HttpInputPurchaseOrderInfo) error {
	// TODO not sure where this comes from really, assume dynamic

	environment := input.Environment
	microserviceID := input.Dolittle.MicroserviceID
	microserviceName := input.Name
	headImage := input.Extra.Headimage
	runtimeImage := input.Extra.Runtimeimage

	microservice := k8s.Microservice{
		ID:          microserviceID,
		Name:        microserviceName,
		Tenant:      customer,
		Application: application,
		Environment: environment,
		ResourceID:  string(tenant),
		Kind:        platform.MicroserviceKindPurchaseOrderAPI,
	}

	ctx := context.TODO()

	// if err := r.rawdatalogRepo.EnsureForPurchaseOrderAPI(namespace, environment, tenant, application, input.Extra.Webhooks); err != nil {
	// 	return err
	// }

	if err := r.k8sResource.Create(namespace, headImage, runtimeImage, microservice, tenant, input.Extra, ctx); err != nil {
		return err
	}

	return nil
}

// Delete stops the running purchase order api and deletes the kubernetes resources.
func (r *repo) Delete(namespace string, microserviceID string) error {
	ctx := context.TODO()
	return r.k8sResource.Delete(namespace, microserviceID, ctx)
}
