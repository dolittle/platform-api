package microservice

import (
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"k8s.io/client-go/kubernetes"
)

type service struct {
	simpleRepo                 simpleRepo
	businessMomentsAdaptorRepo businessMomentsAdaptorRepo
	k8sDolittleRepo            platform.K8sRepo
	gitRepo                    storage.Repo
	k8sClient                  *kubernetes.Clientset
}
