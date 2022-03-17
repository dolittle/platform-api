package customer

import (
	"github.com/dolittle/platform-api/pkg/k8s"
	jobK8s "github.com/dolittle/platform-api/pkg/platform/job/k8s"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type service struct {
	k8sclient         kubernetes.Interface
	storageRepo       storage.RepoCustomer
	jobResourceConfig jobK8s.CreateResourceConfig
	logContext        logrus.FieldLogger
	roleBindingRepo   k8s.RepoRoleBinding
}

func NewService(
	k8sclient kubernetes.Interface,
	storageRepo storage.RepoCustomer,
	jobResourceConfig jobK8s.CreateResourceConfig,
	logContext logrus.FieldLogger,
	roleBindingRepo k8s.RepoRoleBinding,
) service {
	return service{
		k8sclient:         k8sclient,
		storageRepo:       storageRepo,
		jobResourceConfig: jobResourceConfig,
		logContext:        logContext,
		roleBindingRepo:   roleBindingRepo,
	}
}
