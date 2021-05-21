package application

import (
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"k8s.io/client-go/kubernetes"
)

type service struct {
	gitRepo         storage.Repo
	k8sDolittleRepo platform.K8sRepo
	k8sClient       *kubernetes.Clientset
}
