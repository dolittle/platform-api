package microservice

import (
	"context"
	"errors"
	"fmt"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/rawdatalog"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type PurchaseOrderAPIRepo struct {
	k8sClient      *kubernetes.Clientset
	rawDataLogRepo rawdatalog.RawDataLogIngestorRepo
	kind           platform.MicroserviceKind
}

// NewPurchaseOrderAPIRepo creates a new instance of purchaseorderapiRepo.
func NewPurchaseOrderAPIRepo(k8sClient *kubernetes.Clientset, rawDataLogRepo rawdatalog.RawDataLogIngestorRepo) PurchaseOrderAPIRepo {
	return PurchaseOrderAPIRepo{
		k8sClient,
		rawDataLogRepo,
		platform.MicroserviceKindPurchaseOrderAPI,
	}
}

// Create creates a new PurchaseOrderAPI microservice, and a RawDataLog and WebhookListener if they don't exist.
func (r PurchaseOrderAPIRepo) Create(namespace string, tenant k8s.Tenant, application k8s.Application, input platform.HttpInputPurchaseOrderInfo) error {
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
		ResourceID:  TodoCustomersTenantID,
		Kind:        r.kind,
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

	//TODO: Customise the config to adhere to how we create purchase order api
	// TODO: Add webhooks

	// TODO: add rawDataLogMicroserviceID
	// configFiles.Data = map[string]string{}
	// We store the config data into the config-Files for the service to pick up on
	// b, _ := json.MarshalIndent(input, "", "  ")
	// configFiles.Data["microservice_data_from_studio.json"] = string(b)
	// TODO lookup to see if it exists?
	// exists, err := r.rawDataLogRepo.Exists(namespace, environment, microserviceID)
	//exists, err := s.rawDataLogIngestorRepo.Exists(namespace, ms.Environment, ms.Dolittle.MicroserviceID)
	//if err != nil {
	//	// TODO change
	//	utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
	//	return
	//}

	// if !exists {
	// 	fmt.Println("Raw Data Log does not exist")
	// }

	// Assuming the namespace exists

	return nil
}

// Delete stops the running purchase order api and deletes the kubernetes resources.
func (r PurchaseOrderAPIRepo) Delete(namespace string, microserviceID string) error {
	ctx := context.TODO()

	deployment, err := r.stopDeployment(ctx, namespace, microserviceID)
	if err != nil {
		return err
	}

	listOpts := metaV1.ListOptions{
		LabelSelector: labels.FormatLabels(deployment.GetObjectMeta().GetLabels()),
	}

	if err = k8sDeleteConfigmaps(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = k8sDeleteSecrets(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = k8sDeleteServices(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = k8sDeleteDeployment(r.k8sClient, ctx, namespace, &deployment); err != nil {
		return err
	}
	return nil
}

func (r PurchaseOrderAPIRepo) stopDeployment(ctx context.Context, namespace, microserviceID string) (v1.Deployment, error) {
	deployment, err := k8sGetDeployment(r.k8sClient, ctx, namespace, microserviceID)
	if err != nil {
		return deployment, err
	}

	if err = k8sStopDeployment(r.k8sClient, ctx, namespace, &deployment); err != nil {
		return deployment, err
	}
	return deployment, nil
}

func (r PurchaseOrderAPIRepo) createRawDataLogIfNotExists() error {
	return errors.New("Not implemented")
}

func (r PurchaseOrderAPIRepo) createPurchaseOrderMicroservice(namespace, headImage, runtimeImage string, microservice k8s.Microservice, ctx context.Context) error {
	opts := metaV1.CreateOptions{}

	microserviceConfigmap := k8s.NewMicroserviceConfigmap(microservice, TodoCustomersTenantID)
	deployment := k8s.NewDeployment(microservice, headImage, runtimeImage)
	service := k8s.NewService(microservice)
	configEnvVariables := k8s.NewEnvVariablesConfigmap(microservice)
	configFiles := k8s.NewConfigFilesConfigmap(microservice)
	configSecrets := k8s.NewEnvVariablesSecret(microservice)

	r.ModifyEnvironmentVariablesConfigMap(configEnvVariables, microservice)

	// ConfigMaps
	_, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Create(ctx, microserviceConfigmap, opts)
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("microservice config map") }) != nil {
		return err
	}
	_, err = r.k8sClient.CoreV1().ConfigMaps(namespace).Create(ctx, configEnvVariables, opts)
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("config env variables") }) != nil {
		return err
	}
	_, err = r.k8sClient.CoreV1().ConfigMaps(namespace).Create(ctx, configFiles, opts)
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("config files") }) != nil {
		return err
	}
	_, err = r.k8sClient.CoreV1().Secrets(namespace).Create(ctx, configSecrets, opts)
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("config secrets") }) != nil {
		return err
	}
	_, err = r.k8sClient.CoreV1().Services(namespace).Create(ctx, service, opts)
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("service") }) != nil {
		return err
	}
	_, err = r.k8sClient.AppsV1().Deployments(namespace).Create(ctx, deployment, opts)
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("deployment") }) != nil {
		return err
	}

	return nil
}

func (r PurchaseOrderAPIRepo) ModifyEnvironmentVariablesConfigMap(environmentVariablesConfigMap *corev1.ConfigMap, microservice k8s.Microservice) {
	resources := k8s.NewMicroserviceResources(microservice, TodoCustomersTenantID)
	mongoDBURL := resources[TodoCustomersTenantID].Readmodels.Host
	readmodelDBName := resources[TodoCustomersTenantID].Readmodels.Database

	tenantID := TodoCustomersTenantID
	natsClusterURL := fmt.Sprintf("%s-rawdatalogv1-nats.application-%s.svc.cluster.local:4222", microservice.Environment, microservice.Application.ID)

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

func (r PurchaseOrderAPIRepo) createWebhookListenerIfNotExists() error {
	return errors.New("Not implemented")
}

func (r PurchaseOrderAPIRepo) addWebhookEndpoints() error {
	return errors.New("Not implemented")
}
