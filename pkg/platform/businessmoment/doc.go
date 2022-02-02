package businessmoment

import (
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/businessmomentsadaptor"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type service struct {
	logContext            logrus.FieldLogger
	k8sClient             kubernetes.Interface
	gitRepo               storage.Repo
	k8sDolittleRepo       platformK8s.K8sRepo
	k8sBusinessMomentRepo businessmomentsadaptor.Repo
}
