package microservice

import (
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/requesthandler"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
)

type service struct {
	handlers        requesthandler.Handlers
	parser          requesthandler.Parser
	k8sDolittleRepo platform.K8sRepo
	gitRepo         storage.Repo
}
