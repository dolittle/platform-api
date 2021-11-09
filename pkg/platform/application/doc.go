package application

import (
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/storage"
)

type service struct {
	subscriptionID      string
	externalClusterHost string
	gitRepo             storage.Repo
	k8sDolittleRepo     platform.K8sRepo
}
