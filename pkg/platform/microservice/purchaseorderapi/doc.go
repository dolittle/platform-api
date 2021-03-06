package purchaseorderapi

import (
	"context"

	"github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Repo interface {
	// Create creates the microservice by committing it to a persistent storage and applying its kubernetes resources
	Create(namespace string, customer k8s.Tenant, application k8s.Application, customerTenants []platform.CustomerTenantInfo, input platform.HttpInputPurchaseOrderInfo) error
	// Delete deletes the microservice by deleting its kubernetes resources
	Delete(applicationID, environment, microserviceID string) error
	// Exists checks whether a purchase order api with the same name has already been created
	Exists(namespace string, customer k8s.Tenant, application k8s.Application, customerTenants []platform.CustomerTenantInfo, input platform.HttpInputPurchaseOrderInfo) (bool, error)
	// EnvironmentHasPurchaseOrderAPI checks whether the given environment has a purchase order api deployed
	EnvironmentHasPurchaseOrderAPI(namespace string, input platform.HttpInputPurchaseOrderInfo) (bool, error)
}

type K8sResource interface {
	Create(ctx context.Context, namspace, headImage, runtimeImage string, k8sMicroservice k8s.Microservice, customerTenants []platform.CustomerTenantInfo, extra platform.HttpInputPurchaseOrderExtra) error
	Delete(ctx context.Context, applicationID, environment, microserviceID string) error
}
type K8sResourceSpecFactory interface {
	CreateAll(headImage, runtimeImage string, k8sMicroservice k8s.Microservice, customerTenants []platform.CustomerTenantInfo, extra platform.HttpInputPurchaseOrderExtra) K8sResources
}
type K8sResources struct {
	MicroserviceConfigMap *corev1.ConfigMap
	ConfigEnvVariables    *corev1.ConfigMap
	ConfigFiles           *corev1.ConfigMap
	ConfigSecrets         *corev1.Secret
	Service               *corev1.Service
	Deployment            *v1.Deployment
}
