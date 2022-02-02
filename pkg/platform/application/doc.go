package application

import (
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
