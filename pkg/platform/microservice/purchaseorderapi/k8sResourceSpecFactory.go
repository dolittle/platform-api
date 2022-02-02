package purchaseorderapi

import (
	"fmt"
	"strings"

	"github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	corev1 "k8s.io/api/core/v1"
)

type k8sResourceSpecFactory struct{}

// NewRepo creates a new instance of purchaseorderapiRepo.
func NewK8sResourceSpecFactory() K8sResourceSpecFactory {
	return &k8sResourceSpecFactory{}
}

func (r *k8sResourceSpecFactory) CreateAll(headImage, runtimeImage string, k8sMicroservice k8s.Microservice, customerTenants []platform.CustomerTenantInfo, extra platform.HttpInputPurchaseOrderExtra) K8sResources {
	resources := K8sResources{}
	resources.MicroserviceConfigMap = k8s.NewMicroserviceConfigmap(k8sMicroservice, customerTenants)
	resources.Deployment = k8s.NewDeployment(k8sMicroservice, headImage, runtimeImage)
	resources.Service = k8s.NewService(k8sMicroservice)
	resources.ConfigEnvVariables = k8s.NewEnvVariablesConfigmap(k8sMicroservice)
	resources.ConfigFiles = k8s.NewConfigFilesConfigmap(k8sMicroservice)
	resources.ConfigSecrets = k8s.NewEnvVariablesSecret(k8sMicroservice)
	r.modifyEnvironmentVariablesConfigMap(resources.ConfigEnvVariables, k8sMicroservice, customerTenants, extra)
	return resources
}

func (r *k8sResourceSpecFactory) modifyEnvironmentVariablesConfigMap(environmentVariablesConfigMap *corev1.ConfigMap, k8sMicroservice k8s.Microservice, customerTenants []platform.CustomerTenantInfo, extra platform.HttpInputPurchaseOrderExtra) {
	// TODO I think we need to come back and look at this properly
	// I am not 100% what we are thinking, as if we rely on an ingress we need many webhooks
	// I believe the issue is we want the tenantID to come from the payload
	tenantID := customerTenants[0].CustomerTenantID
	resources := k8s.NewMicroserviceResources(k8sMicroservice, customerTenants)
	mongoDBURL := resources[tenantID].Readmodels.Host
	readmodelDBName := resources[tenantID].Readmodels.Database

	natsClusterURL := fmt.Sprintf("%s-nats.application-%s.svc.cluster.local:4222", strings.ToLower(k8sMicroservice.Environment), k8sMicroservice.Application.ID)

	environmentVariablesConfigMap.Data = map[string]string{
		"LOG_LEVEL":                 "debug",
		"DATABASE_READMODELS_URL":   mongoDBURL,
		"DATABASE_READMODELS_NAME":  readmodelDBName,
		"NODE_ENV":                  "production",
		"TENANT":                    tenantID,
		"SERVER_PORT":               "80",
		"NATS_CLUSTER_URL":          natsClusterURL,
		"NATS_START_FROM_BEGINNING": "false",
		"LOG_OUTPUT_FORMAT":         "json",
	}
}
