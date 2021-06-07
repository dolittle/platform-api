package storage

import (
	"errors"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
)

// TODO need a better name
// Interface for writing data for the automation part of the platform
type Repo interface {
	SaveTerraformApplication(application platform.TerraformApplication) error
	GetTerraformApplication(tenantID string, applicationID string) (platform.TerraformApplication, error)

	SaveTerraformTenant(tenant platform.TerraformCustomer) error
	GetTerraformTenant(tenantID string) (platform.TerraformCustomer, error)

	GetApplication(tenantID string, applicationID string) (platform.HttpResponseApplication, error)
	SaveApplication(application platform.HttpResponseApplication) error
	GetApplications(tenantID string) ([]platform.HttpResponseApplication, error)

	SaveMicroservice(tenantID string, applicationID string, environment string, microserviceID string, data []byte) error
	GetMicroservice(tenantID string, applicationID string, environment string, microserviceID string) ([]byte, error)
	DeleteMicroservice(tenantID string, applicationID string, environment string, microserviceID string) error
	GetMicroservices(tenantID string, applicationID string) ([]platform.HttpMicroserviceBase, error)

	SaveStudioConfig(tenantID string, config platform.StudioConfig) error
	GetStudioConfig(tenantID string) (platform.StudioConfig, error)
	IsAutomationEnabled(tenantID string, applicationID string, environment string) bool
	CheckAutomationEnabledViaCustomer(customer platform.StudioConfig, applicationID string, environment string) bool

	SaveBusinessMoment(tenantID string, input platform.HttpInputBusinessMoment) error
	GetBusinessMoments(tenantID string, applicationID string, environment string) (platform.HttpResponseBusinessMoments, error)
	DeleteBusinessMoment(tenantID string, applicationID string, environment string, microserviceID string, momentID string) error

	SaveBusinessMomentEntity(tenantID string, input platform.HttpInputBusinessMomentEntity) error
	DeleteBusinessMomentEntity(tenantID string, applicationID string, environment string, microserviceID string, entityID string) error
}

var (
	ErrNotFound                  = errors.New("not-found")
	ErrNotBusinessMomentsAdaptor = errors.New("not-business-moments-adaptor")
)
