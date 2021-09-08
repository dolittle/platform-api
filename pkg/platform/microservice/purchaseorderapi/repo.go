package purchaseorderapi

import (
	"context"
	"errors"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice"
	"k8s.io/client-go/kubernetes"
)

type repo struct {
	k8sClient   *kubernetes.Clientset
	k8sResource K8sResource
}

// NewRepo creates a new instance of purchaseorderapiRepo.
func NewRepo(k8sClient *kubernetes.Clientset, k8sResource K8sResource) Repo {
	return &repo{
		k8sClient,
		k8sResource,
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
		ResourceID:  microservice.TodoCustomersTenantID,
		Kind:        platform.MicroserviceKindPurchaseOrderAPI,
	}

	ctx := context.TODO()

	if err := r.createRawDataLogIfNotExists(); err != nil {
		return err
	}

	if err := r.createWebhookListenerIfNotExists(); err != nil {
		return err
	}

	if err := r.k8sResource.Create(namespace, headImage, runtimeImage, microservice, ctx); err != nil {
		return err
	}

	if err := r.addWebhookEndpoints(); err != nil {
		return err
	}

	return nil
}

// Delete stops the running purchase order api and deletes the kubernetes resources.
func (r *repo) Delete(namespace string, microserviceID string) error {
	ctx := context.TODO()
	return r.k8sResource.Delete(namespace, microserviceID, ctx)
}
func (r *repo) createRawDataLogIfNotExists() error {
	return errors.New("Not implemented")
}

func (r *repo) createWebhookListenerIfNotExists() error {
	return errors.New("Not implemented")
}

func (r *repo) addWebhookEndpoints() error {
	return errors.New("Not implemented")
}
