package application

import (
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"k8s.io/client-go/kubernetes"
)

// Environment
// DomainPrefix
type HttpInputApplication struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	TenantID string `json:"tenantId"`
}

type HttpInputEnvironment struct {
	Name         string `json:"name"`
	DomainPrefix string `json:"domainPrefix"`
}

type Application struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	TenantID     string                 `json:"tenantId"`
	Environments []HttpInputEnvironment `json:"environments"`
}

type Storage interface {
	Write(tenantID string, applicationID string, data []byte) error
	Read(tenantID string, applicationID string) ([]byte, error)
	GetAll(tenantID string) ([]Application, error)
}

type service struct {
	storage         Storage
	k8sDolittleRepo platform.K8sRepo
	k8sClient       *kubernetes.Clientset
}

type HttpInput interface{}
