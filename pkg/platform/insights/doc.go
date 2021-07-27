package insights

import (
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/sirupsen/logrus"
)

type service struct {
	logContext      logrus.FieldLogger
	k8sDolittleRepo platform.K8sRepo
}
