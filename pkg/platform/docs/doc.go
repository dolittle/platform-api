package docs

import (
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
)

type service struct {
	gitRepo         storage.Repo
	k8sDolittleRepo platform.K8sRepo
}
