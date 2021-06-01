package businessmomentsadaptor

import (
	"errors"

	coreV1 "k8s.io/api/core/v1"
)

type Repo interface {
	SaveBusinessMomentsConfigmap(newConfigmap coreV1.ConfigMap, data []byte) error
	GetBusinessMomentsConfigmap(applicationID string, environment string, microserviceID string) (coreV1.ConfigMap, error)
}

var (
	ErrNotFound = errors.New("not-found")
)
