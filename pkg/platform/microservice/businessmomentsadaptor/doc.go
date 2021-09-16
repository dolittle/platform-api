package businessmomentsadaptor

import (
	"errors"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	coreV1 "k8s.io/api/core/v1"
)

type K8sRepoRepo interface {
	SaveBusinessMomentsConfigmap(newConfigmap coreV1.ConfigMap, data []byte) error
	GetBusinessMomentsConfigmap(applicationID string, environment string, microserviceID string) (coreV1.ConfigMap, error)
}
type Repo interface {
	Create(namespace string, customer k8s.Tenant, application k8s.Application, applicationIngress k8s.Ingress, tenant platform.TenantId, input platform.HttpInputBusinessMomentAdaptorInfo) error
	Delete(namespace string, microserviceID string) error
}

var (
	ErrNotFound = errors.New("not-found")
)
