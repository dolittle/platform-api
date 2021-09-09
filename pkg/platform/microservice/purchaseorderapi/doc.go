package purchaseorderapi

import (
	"context"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Repo interface {
	// Create creates the microservice by committing it to a persistent storage and applying its kubernetes resources
	Create(namespace string, tenant k8s.Tenant, application k8s.Application, input platform.HttpInputPurchaseOrderInfo) error
	// Delete deletes the microservice by deleting its kubernetes resources
	Delete(namespace, microserviceID string) error
}

type K8sResource interface {
	Create(namspace, headImage, runtimeImage string, k8sMicroservice k8s.Microservice, extra platform.HttpInputPurchaseOrderExtra, ctx context.Context) error
	Delete(namespace, microserviceID string, ctx context.Context) error
}
type K8sResourceSpecFactory interface {
	CreateAll(headImage, runtimeImage string, k8sMicroservice k8s.Microservice, extra platform.HttpInputPurchaseOrderExtra) K8sResources
}
type K8sResources struct {
	MicroserviceConfigMap *corev1.ConfigMap
	ConfigEnvVariables    *corev1.ConfigMap
	ConfigFiles           *corev1.ConfigMap
	ConfigSecrets         *corev1.Secret
	Service               *corev1.Service
	Deployment            *v1.Deployment
}
