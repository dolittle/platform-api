package insights

import (
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/sirupsen/logrus"
)

type service struct {
	logContext      logrus.FieldLogger
	k8sDolittleRepo platform.K8sRepo
	lokiHost        string
}
