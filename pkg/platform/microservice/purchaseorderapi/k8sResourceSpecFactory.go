package purchaseorderapi

import (
	"fmt"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice"
	corev1 "k8s.io/api/core/v1"
)

type k8sResourceSpecFactory struct{}

// NewRepo creates a new instance of purchaseorderapiRepo.
func NewK8sResourceSpecFactory() K8sResourceSpecFactory {
	return &k8sResourceSpecFactory{}
}

func (r *k8sResourceSpecFactory) CreateAll(headImage, runtimeImage string, k8sMicroservice k8s.Microservice) K8sResources {
	resources := K8sResources{}
	resources.MicroserviceConfigMap = k8s.NewMicroserviceConfigmap(k8sMicroservice, microservice.TodoCustomersTenantID)
	resources.Deployment = k8s.NewDeployment(k8sMicroservice, headImage, runtimeImage)
	resources.Service = k8s.NewService(k8sMicroservice)
	resources.ConfigEnvVariables = k8s.NewEnvVariablesConfigmap(k8sMicroservice)
	resources.ConfigFiles = k8s.NewConfigFilesConfigmap(k8sMicroservice)
	resources.ConfigSecrets = k8s.NewEnvVariablesSecret(k8sMicroservice)
	r.modifyEnvironmentVariablesConfigMap(resources.ConfigEnvVariables, k8sMicroservice)

	return resources
}

func (r *k8sResourceSpecFactory) modifyEnvironmentVariablesConfigMap(environmentVariablesConfigMap *corev1.ConfigMap, k8sMicroservice k8s.Microservice) {
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
