package storage

import (
	"errors"

	"github.com/dolittle/platform-api/pkg/platform"
)

// TODO need a better name
// Interface for writing data for the automation part of the platform

type RepoCustomer interface {
	GetCustomers() ([]platform.Customer, error)
	SaveCustomer(customer JSONCustomer) error
}
type RepoCustomerTenants interface {
	GetCustomerTenants(application JSONApplication2) []platform.CustomerTenantInfo
	GetCustomerTenantsByEnvironment(application JSONApplication2, environment string) []platform.CustomerTenantInfo
}

type RepoMicroservice interface {
	SaveMicroservice(tenantID string, applicationID string, environment string, microserviceID string, data interface{}) error
	GetMicroservice(tenantID string, applicationID string, environment string, microserviceID string) ([]byte, error)
	DeleteMicroservice(tenantID string, applicationID string, environment string, microserviceID string) error
	GetMicroservices(tenantID string, applicationID string) ([]platform.HttpMicroserviceBase, error)
}

type RepoApplication interface {
	GetApplication(tenantID string, applicationID string) (JSONApplication2, error)
	SaveApplication2(application JSONApplication2) error
	GetApplications(tenantID string) ([]JSONApplication2, error)
	SaveApplication(application platform.HttpResponseApplication) error
}

type Repo interface {
	RepoMicroservice
	RepoApplication

	SaveTerraformApplication(application platform.TerraformApplication) error
	GetTerraformApplication(tenantID string, applicationID string) (platform.TerraformApplication, error)

	SaveTerraformTenant(tenant platform.TerraformCustomer) error
	GetTerraformTenant(tenantID string) (platform.TerraformCustomer, error)

	SaveStudioConfig(tenantID string, config platform.StudioConfig) error
	GetStudioConfig(tenantID string) (platform.StudioConfig, error)
	IsAutomationEnabledWithStudioConfig(studioConfig platform.StudioConfig, applicationID string, environment string) bool

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
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	TenantID     string            `json:"tenantId"`
	TenantName   string            `json:"tenantName"`
	Environments []JSONEnvironment `json:"environments"`
}

// JSONEnvironment represents the "environments" property in application.json file
type JSONEnvironment struct {
	Name          string                        `json:"name"`
	TenantID      string                        `json:"tenantId"`
	ApplicationID string                        `json:"applicationId"`
	Tenants       []platform.TenantId           `json:"tenants"`
	Ingresses     platform.EnvironmentIngresses `json:"ingresses"`
}

type JSONApplication2 struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	TenantID     string             `json:"tenantId"`
	TenantName   string             `json:"tenantName"`
	Environments []JSONEnvironment2 `json:"environments"`
	// TODO do we fire an event around?
	// TODO do we update rest?
	// TODO do we update the repo and signal platform-api to refresh?
	//JobState     JSONApplicationJobState `json:"jobState"`
	Status JSONBuildStatus `json:"status"`
}

type JSONApplicationJobState struct {
	State int `json:"state"`
	ID    int `json:"id"`
}

// JSONEnvironment represents the "environments" property in application.json file
type JSONEnvironment2 struct {
	Name                  string                        `json:"name"`
	CustomerTenants       []platform.CustomerTenantInfo `json:"customerTenants"`
	WelcomeMicroserviceID string                        `json:"welcomeMicroserviceID"`
}

type JSONCustomer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
