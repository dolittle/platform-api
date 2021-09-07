package microservice

import (
	"context"
	"errors"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/rawdatalog"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type purchaseOrderAPIRepo struct {
	k8sClient      *kubernetes.Clientset
	rawDataLogRepo rawdatalog.RawDataLogIngestorRepo
	kind           platform.MicroserviceKind
}

// Creates a new instance of purchaseorderapiRepo
func NewPurchaseOrderAPIRepo(k8sClient *kubernetes.Clientset, rawDataLogRepo rawdatalog.RawDataLogIngestorRepo) purchaseOrderAPIRepo {
	return purchaseOrderAPIRepo{
		k8sClient,
		rawDataLogRepo,
		platform.MicroserviceKindPurchaseOrderAPI,
	}
}

func (r purchaseOrderAPIRepo) Create(namespace string, tenant k8s.Tenant, application k8s.Application, input platform.HttpInputPurchaseOrderInfo) error {
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
		ResourceID:  todoCustomersTenantID,
		Kind:        r.kind,
	}
	microserviceConfigmap := k8s.NewMicroserviceConfigmap(microservice, todoCustomersTenantID)
	deployment := k8s.NewDeployment(microservice, headImage, runtimeImage)
	service := k8s.NewService(microservice)
	configEnvVariables := k8s.NewEnvVariablesConfigmap(microservice)
	configFiles := k8s.NewConfigFilesConfigmap(microservice)
	configSecrets := k8s.NewEnvVariablesSecret(microservice)
	// TODO: Add webhooks

	// TODO: add rawDataLogMicroserviceID
	// configFiles.Data = map[string]string{}
	// We store the config data into the config-Files for the service to pick up on
	// b, _ := json.MarshalIndent(input, "", "  ")
	// configFiles.Data["microservice_data_from_studio.json"] = string(b)

	if err := createRawDataLogIfNotExists(); err != nil {
		return err
	}
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
	var err error
	ctx := context.TODO()

	opts := metaV1.CreateOptions{}

	// ConfigMaps
	_, err = r.k8sClient.CoreV1().ConfigMaps(namespace).Create(ctx, microserviceConfigmap, opts)
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

func (r purchaseOrderAPIRepo) Delete(namespace string, microserviceID string) error {
	ctx := context.TODO()

	deployment, err := k8sGetDeployment(r.k8sClient, ctx, namespace, microserviceID)
	if err != nil {
		return err
	}

	if err = k8sStopDeployment(r.k8sClient, ctx, namespace, &deployment); err != nil {
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

func createRawDataLogIfNotExists() error {
	return errors.New("Not implemented")
}
