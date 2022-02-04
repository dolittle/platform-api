package insights

import (
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/sirupsen/logrus"
)

type service struct {
	logContext      logrus.FieldLogger
	k8sDolittleRepo platformK8s.K8sRepo
	lokiHost        string
}
