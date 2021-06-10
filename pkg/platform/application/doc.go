package application

import (
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
)

type service struct {
	subscriptionID  string
	gitRepo         storage.Repo
	k8sDolittleRepo platform.K8sRepo
}
