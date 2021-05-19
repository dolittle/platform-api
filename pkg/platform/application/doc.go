package application

import (
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"k8s.io/client-go/kubernetes"
)

type Storage interface {
	Write(tenantID string, applicationID string, data []byte) error
	Read(tenantID string, applicationID string) ([]byte, error)
	GetAll(tenantID string) ([]platform.HttpResponseApplication, error)
}

type service struct {
	gitRepo         *gitRepo
	k8sDolittleRepo platform.K8sRepo
	k8sClient       *kubernetes.Clientset
}
