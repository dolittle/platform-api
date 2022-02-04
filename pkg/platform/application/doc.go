package application

import (
	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/simple"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type service struct {
	subscriptionID          string
	externalClusterHost     string
	simpleRepo              simple.Repo
	gitRepo                 storage.Repo
	k8sDolittleRepo         platformK8s.K8sRepo
	k8sClient               kubernetes.Interface
	platformOperationsImage string
	platformEnvironment     string
	isProduction            bool
	logContext              logrus.FieldLogger
}

type HttpResponseApplication struct {
	ID            string                          `json:"id"`
	Name          string                          `json:"name"`
	TenantID      string                          `json:"tenantId"`
	TenantName    string                          `json:"tenantName"`
	Environments  []HttpResponseEnvironment       `json:"environments"`
	Microservices []platform.HttpMicroserviceBase `json:"microservices,omitempty"`
}

type HttpResponseEnvironment struct {
	AutomationEnabled bool   `json:"automationEnabled"`
	Name              string `json:"name"`
}
