package purchaseorderapi

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type repo struct {
	k8sClient *kubernetes.Clientset
}

// NewRepo creates a new instance of purchaseorderapiRepo.
func NewRepo(k8sClient *kubernetes.Clientset) Repo {
	return &repo{
		k8sClient,
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
		Kind:        MicroserviceKindPurchaseOrderAPI,
	}

	ctx := context.TODO()

	if err := r.createRawDataLogIfNotExists(); err != nil {
		return err
	}

	if err := r.createWebhookListenerIfNotExists(); err != nil {
		return err
	}

	if err := r.createPurchaseOrderMicroservice(namespace, headImage, runtimeImage, microservice, ctx); err != nil {
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

	deployment, err := r.getAndStopDeployment(ctx, namespace, microserviceID)
	if err != nil {
		return err
	}

	return r.deleteResources(ctx, namespace, deployment)
}

func (r *repo) getAndStopDeployment(ctx context.Context, namespace, microserviceID string) (v1.Deployment, error) {
	deployment, err := microservice.K8sGetDeployment(r.k8sClient, ctx, namespace, microserviceID)
	if err != nil {
		return deployment, err
	}

	if err = microservice.K8sStopDeployment(r.k8sClient, ctx, namespace, &deployment); err != nil {
		return deployment, err
	}
	return deployment, nil
}

func (r *repo) deleteResources(ctx context.Context, namespace string, deployment v1.Deployment) error {
	listOpts := metaV1.ListOptions{
		LabelSelector: labels.FormatLabels(deployment.GetObjectMeta().GetLabels()),
	}
	var err error
	if err = microservice.K8sDeleteConfigmaps(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = microservice.K8sDeleteSecrets(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = microservice.K8sDeleteServices(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = microservice.K8sDeleteDeployment(r.k8sClient, ctx, namespace, &deployment); err != nil {
		return err
	}
	return nil
}

func (r *repo) createRawDataLogIfNotExists() error {
	return errors.New("Not implemented")
}

func (r *repo) createPurchaseOrderMicroservice(namespace, headImage, runtimeImage string, k8sMicroservice k8s.Microservice, ctx context.Context) error {
	opts := metaV1.CreateOptions{}

	microserviceConfigmap := k8s.NewMicroserviceConfigmap(k8sMicroservice, microservice.TodoCustomersTenantID)
	deployment := k8s.NewDeployment(k8sMicroservice, headImage, runtimeImage)
	service := k8s.NewService(k8sMicroservice)
	configEnvVariables := k8s.NewEnvVariablesConfigmap(k8sMicroservice)
	configFiles := k8s.NewConfigFilesConfigmap(k8sMicroservice)
	configSecrets := k8s.NewEnvVariablesSecret(k8sMicroservice)

	r.ModifyEnvironmentVariablesConfigMap(configEnvVariables, k8sMicroservice)

	// ConfigMaps
	_, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Create(ctx, microserviceConfigmap, opts)
	if microservice.K8sHandleResourceCreationError(err, func() { microservice.K8sPrintAlreadyExists("microservice config map") }) != nil {
		return err
	}
	_, err = r.k8sClient.CoreV1().ConfigMaps(namespace).Create(ctx, configEnvVariables, opts)
	if microservice.K8sHandleResourceCreationError(err, func() { microservice.K8sPrintAlreadyExists("config env variables") }) != nil {
		return err
	}
	_, err = r.k8sClient.CoreV1().ConfigMaps(namespace).Create(ctx, configFiles, opts)
	if microservice.K8sHandleResourceCreationError(err, func() { microservice.K8sPrintAlreadyExists("config files") }) != nil {
		return err
	}
	_, err = r.k8sClient.CoreV1().Secrets(namespace).Create(ctx, configSecrets, opts)
	if microservice.K8sHandleResourceCreationError(err, func() { microservice.K8sPrintAlreadyExists("config secrets") }) != nil {
		return err
	}
	_, err = r.k8sClient.CoreV1().Services(namespace).Create(ctx, service, opts)
	if microservice.K8sHandleResourceCreationError(err, func() { microservice.K8sPrintAlreadyExists("service") }) != nil {
		return err
	}
	_, err = r.k8sClient.AppsV1().Deployments(namespace).Create(ctx, deployment, opts)
	if microservice.K8sHandleResourceCreationError(err, func() { microservice.K8sPrintAlreadyExists("deployment") }) != nil {
		return err
	}

	return nil
}

func (r *repo) ModifyEnvironmentVariablesConfigMap(environmentVariablesConfigMap *corev1.ConfigMap, k8sMicroservice k8s.Microservice) {
	resources := k8s.NewMicroserviceResources(k8sMicroservice, microservice.TodoCustomersTenantID)
	mongoDBURL := resources[microservice.TodoCustomersTenantID].Readmodels.Host
	readmodelDBName := resources[microservice.TodoCustomersTenantID].Readmodels.Database

	tenantID := microservice.TodoCustomersTenantID
	natsClusterURL := fmt.Sprintf("%s-rawdatalogv1-nats.application-%s.svc.cluster.local:4222", strings.ToLower(k8sMicroservice.Environment), k8sMicroservice.Application.ID)

	environmentVariablesConfigMap.Data = map[string]string{
		"LOG_LEVEL":                 "debug",
		"DATABASE_READMODELS_URL":   mongoDBURL,
		"DATABASE_READMODELS_NAME":  readmodelDBName,
		"NODE_ENV":                  "production",
		"TENANT":                    tenantID,
		"SERVER_PORT":               "8080",
		"NATS_CLUSTER_URL":          natsClusterURL,
		"NATS_START_FROM_BEGINNING": "false",
		"LOG_OUTPUT_FORMAT":         "json",
	}
}

func (r *repo) createWebhookListenerIfNotExists() error {
	return errors.New("Not implemented")
}

func (r *repo) addWebhookEndpoints() error {
	return errors.New("Not implemented")
}
