package application

import (
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/storage"
)

type service struct {
	subscriptionID string
	// externalClusterHost used to signify the full url to the apiserver outside of the cluster
	externalClusterHost string
	gitRepo             storage.Repo
	k8sDolittleRepo     platform.K8sRepo
}
