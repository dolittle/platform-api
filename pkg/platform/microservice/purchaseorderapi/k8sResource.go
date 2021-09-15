package purchaseorderapi

import (
	"context"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	microserviceK8s "github.com/dolittle-entropy/platform-api/pkg/platform/microservice/k8s"
	v1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type k8sResource struct {
	k8sClient   kubernetes.Interface
	specFactory K8sResourceSpecFactory
}

// NewRepo creates a new instance of purchaseorderapiRepo.
func NewK8sResource(k8sClient kubernetes.Interface, specFactory K8sResourceSpecFactory) K8sResource {
	return &k8sResource{
		k8sClient,
		specFactory,
	}
}

// Create creates a new PurchaseOrderAPI microservice, and a RawDataLog and WebhookListener if they don't exist.
func (r *k8sResource) Create(namespace, headImage, runtimeImage string, k8sMicroservice k8s.Microservice, extra platform.HttpInputPurchaseOrderExtra, ctx context.Context) error {
	opts := metaV1.CreateOptions{}

	resources := r.specFactory.CreateAll(headImage, runtimeImage, k8sMicroservice, extra)

	// ConfigMaps
	_, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Create(ctx, resources.MicroserviceConfigMap, opts)
	if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("microservice config map") }) != nil {
		return err
	}
	_, err = r.k8sClient.CoreV1().ConfigMaps(namespace).Create(ctx, resources.ConfigEnvVariables, opts)
	if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("config env variables") }) != nil {
		return err
	}
	_, err = r.k8sClient.CoreV1().ConfigMaps(namespace).Create(ctx, resources.ConfigFiles, opts)
	if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("config files") }) != nil {
		return err
	}
	_, err = r.k8sClient.CoreV1().Secrets(namespace).Create(ctx, resources.ConfigSecrets, opts)
	if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("config secrets") }) != nil {
		return err
	}
	_, err = r.k8sClient.CoreV1().Services(namespace).Create(ctx, resources.Service, opts)
	if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("service") }) != nil {
		return err
	}
	_, err = r.k8sClient.AppsV1().Deployments(namespace).Create(ctx, resources.Deployment, opts)
	if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("deployment") }) != nil {
		return err
	}

	return nil
}

// Delete stops the running purchase order api and deletes the kubernetes resources.
func (r *k8sResource) Delete(namespace, microserviceID string, ctx context.Context) error {
	deployment, err := r.getAndStopDeployment(ctx, namespace, microserviceID)
	if err != nil {
		return err
	}

	return r.deleteResources(ctx, namespace, deployment)
}

func (r *k8sResource) getAndStopDeployment(ctx context.Context, namespace, microserviceID string) (v1.Deployment, error) {
	deployment, err := microserviceK8s.K8sGetDeployment(r.k8sClient, ctx, namespace, microserviceID)
	if err != nil {
		return deployment, err
	}

	if err = microserviceK8s.K8sStopDeployment(r.k8sClient, ctx, namespace, &deployment); err != nil {
		return deployment, err
	}
	return deployment, nil
}

func (r *k8sResource) deleteResources(ctx context.Context, namespace string, deployment v1.Deployment) error {
	listOpts := metaV1.ListOptions{
		LabelSelector: labels.FormatLabels(deployment.GetObjectMeta().GetLabels()),
	}
	var err error
	if err = microserviceK8s.K8sDeleteConfigmaps(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = microserviceK8s.K8sDeleteSecrets(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = microserviceK8s.K8sDeleteServices(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = microserviceK8s.K8sDeleteDeployment(r.k8sClient, ctx, namespace, &deployment); err != nil {
		return err
	}
	return nil
}
