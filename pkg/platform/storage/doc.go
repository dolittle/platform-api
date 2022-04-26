package storage

import (
	"errors"

	"github.com/dolittle/platform-api/pkg/platform"
)

type RepoCustomer interface {
	GetCustomers() ([]platform.Customer, error)
	SaveCustomer(customer JSONCustomer) error
}
type RepoCustomerTenants interface {
	GetCustomerTenants(application JSONApplication) []platform.CustomerTenantInfo
	GetCustomerTenantsByEnvironment(application JSONApplication, environment string) []platform.CustomerTenantInfo
}

type RepoMicroservice interface {
	SaveMicroservice(customerID string, applicationID string, environment string, microserviceID string, data interface{}) error
	GetMicroservice(customerID string, applicationID string, environment string, microserviceID string) ([]byte, error)
	DeleteMicroservice(customerID string, applicationID string, environment string, microserviceID string) error
	GetMicroservices(customerID string, applicationID string) ([]platform.HttpMicroserviceBase, error)
}

type RepoApplication interface {
	GetApplication(customerID string, applicationID string) (JSONApplication, error)
	SaveApplication(application JSONApplication) error
	GetApplications(customerID string) ([]JSONApplication, error)
}

type Repo interface {
	RepoMicroservice
	RepoApplication

	SaveTerraformApplication(application platform.TerraformApplication) error
	GetTerraformApplication(customerID string, applicationID string) (platform.TerraformApplication, error)

	SaveTerraformTenant(tenant platform.TerraformCustomer) error
	GetTerraformTenant(customerID string) (platform.TerraformCustomer, error)

	SaveStudioConfig(customerID string, config platform.StudioConfig) error
	GetStudioConfig(customerID string) (platform.StudioConfig, error)
	IsAutomationEnabledWithStudioConfig(studioConfig platform.StudioConfig, applicationID string, environment string) bool

	SaveBusinessMoment(customerID string, input platform.HttpInputBusinessMoment) error
	GetBusinessMoments(customerID string, applicationID string, environment string) (platform.HttpResponseBusinessMoments, error)
	DeleteBusinessMoment(customerID string, applicationID string, environment string, microserviceID string, momentID string) error

	SaveBusinessMomentEntity(customerID string, input platform.HttpInputBusinessMomentEntity) error
	DeleteBusinessMomentEntity(customerID string, applicationID string, environment string, microserviceID string, entityID string) error
}

var (
	ErrNotFound                  = errors.New("not-found")
	ErrNotBusinessMomentsAdaptor = errors.New("not-business-moments-adaptor")
)

// JSONApplication represents the application.json file

const (
	BuildStatusStateWaiting         = "waiting"
	BuildStatusStatePending         = "building"
	BuildStatusStateFinishedSuccess = "finished:success"
	BuildStatusStateFinishedFailed  = "finished:failed"
)

type JSONBuildStatus struct {
	State      string `json:"status"`
	StartedAt  string `json:"startedAt"`
	FinishedAt string `json:"finishedAt"`
}

type JSONApplication struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	CustomerID   string `json:"customerId"`
	CustomerName string `json:"customerName"`

	Environments []JSONEnvironment `json:"environments"`
	Status       JSONBuildStatus   `json:"status"`
}

// TODO I wonder where best to store this
type JSONEnvironmentConnections struct {
	Kafka       bool `json:"kafka"`
	M3Connector bool `json:"m3Connector"`
}

type JSONEnvironment struct {
	Name                  string                        `json:"name"`
	CustomerTenants       []platform.CustomerTenantInfo `json:"customerTenants"`
	WelcomeMicroserviceID string                        `json:"welcomeMicroserviceID"`
	Connections           JSONEnvironmentConnections    `json:"connections"`
}

type JSONCustomer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
