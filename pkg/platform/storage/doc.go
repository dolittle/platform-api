package storage

import (
	"github.com/dolittle-entropy/platform-api/pkg/platform"
)

// TODO need a better name
// Interface for writing data for the automation part of the platform
type Repo interface {
	SaveTenant(tenant platform.TerraformCustomer) error
	GetTenant(tenantID string) (platform.TerraformCustomer, error)

	GetApplication(tenantID string, applicationID string) (platform.HttpResponseApplication, error)
	SaveApplication(application platform.HttpResponseApplication) error
	GetApplications(tenantID string) ([]platform.HttpResponseApplication, error)

	SaveMicroservice(tenantID string, applicationID string, environment string, microserviceID string, data []byte) error
	GetMicroservice(tenantID string, applicationID string, environment string, microserviceID string) ([]byte, error)
	GetMicroservices(tenantID string, applicationID string) ([]platform.HttpMicroserviceBase, error)

	GetStudioConfig(tenantID string) (platform.StudioConfig, error)
	IsAutomationEnabled(tenantID string, applicationID string, environment string) bool
	CheckAutomationEnabledViaCustomer(customer platform.StudioConfig, applicationID string, environment string) bool
}
