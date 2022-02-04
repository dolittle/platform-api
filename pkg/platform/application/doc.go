package application

import (
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/sirupsen/logrus"
)

type service struct {
	subscriptionID string
	// externalClusterHost used to signify the full url to the apiserver outside of the cluster
	externalClusterHost string
	gitRepo             storage.Repo
	k8sDolittleRepo     platformK8s.K8sRepo
	logContext          logrus.FieldLogger
}
