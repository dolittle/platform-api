package businessmoment

import (
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/businessmomentsadaptor"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type service struct {
	logContext            logrus.FieldLogger
	k8sClient             kubernetes.Interface
	gitRepo               storage.Repo
	k8sDolittleRepo       platform.K8sRepo
	k8sBusinessMomentRepo businessmomentsadaptor.Repo
}
