package microservice

import (
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
)

type service struct {
	simpleRepo                 simpleRepo
	businessMomentsAdaptorRepo businessMomentsAdaptorRepo
	k8sDolittleRepo            platform.K8sRepo
	gitRepo                    storage.Repo
}
