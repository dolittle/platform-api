package storage

import "github.com/dolittle-entropy/platform-api/pkg/platform"

type Repo interface {
	GetApplication(tenantID string, applicationID string) (platform.HttpResponseApplication, error)
	SaveApplication(application platform.HttpResponseApplication) error
	GetApplications(tenantID string) ([]platform.HttpResponseApplication, error)

	SaveMicroservice(tenantID string, applicationID string, environment string, microserviceID string, data []byte) error
	GetMicroservice(tenantID string, applicationID string, environment string, microserviceID string) ([]byte, error)
	GetMicroservices(tenantID string, applicationID string) ([]platform.HttpMicroserviceBase, error)
}
