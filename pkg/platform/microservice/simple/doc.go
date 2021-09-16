package simple

import (
	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
)

type Repo interface {
	// Create creates the microservice by committing it to a persistent storage and applying its kubernetes resources
	Create(namespace string, customer k8s.Tenant, application k8s.Application, ingress k8s.Ingress, tenant platform.TenantId, input platform.HttpInputSimpleInfo) error
	// Delete deletes the microservice by deleting its kubernetes resources
	Delete(namespace, microserviceID string) error
}
