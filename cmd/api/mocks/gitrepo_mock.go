package mocks

import (
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/storage"
)

type GitStorageRepoMock struct {
}

func (m GitStorageRepoMock) SaveMicroservice(customerID string, applicationID string, environment string, microserviceID string, data interface{}) error {
	return nil
}

func (m GitStorageRepoMock) GetMicroservice(customerID string, applicationID string, environment string, microserviceID string) ([]byte, error) {
	return nil, nil
}

func (m GitStorageRepoMock) DeleteMicroservice(customerID string, applicationID string, environment string, microserviceID string) error {
	return nil
}

func (m GitStorageRepoMock) GetMicroservices(customerID string, applicationID string) ([]platform.HttpMicroserviceBase, error) {
	return nil, nil
}

func (m GitStorageRepoMock) GetApplication(customerID string, applicationID string) (storage.JSONApplication, error) {
	return storage.JSONApplication{}, nil
}

func (m GitStorageRepoMock) SaveApplication(application storage.JSONApplication) error {
	return nil
}

func (m GitStorageRepoMock) GetApplications(customerID string) ([]storage.JSONApplication, error) {
	return nil, nil
}

func (m GitStorageRepoMock) SaveTerraformApplication(application platform.TerraformApplication) error {
	return nil
}

func (m GitStorageRepoMock) GetTerraformApplication(customerID string, applicationID string) (platform.TerraformApplication, error) {
	return platform.TerraformApplication{}, nil
}

func (m GitStorageRepoMock) SaveTerraformTenant(tenant platform.TerraformCustomer) error {
	return nil
}

func (m GitStorageRepoMock) GetTerraformTenant(customerID string) (platform.TerraformCustomer, error) {
	return platform.TerraformCustomer{ContainerRegistryName: "test1"}, nil
}

func (m GitStorageRepoMock) SaveStudioConfig(customerID string, config platform.StudioConfig) error {
	return nil
}

func (m GitStorageRepoMock) GetStudioConfig(customerID string) (platform.StudioConfig, error) {
	return platform.StudioConfig{}, nil
}

func (m GitStorageRepoMock) IsAutomationEnabledWithStudioConfig(studioConfig platform.StudioConfig, applicationID string, environment string) bool {
	return false
}

func (m GitStorageRepoMock) SaveBusinessMoment(customerID string, input platform.HttpInputBusinessMoment) error {
	return nil
}

func (m GitStorageRepoMock) GetBusinessMoments(customerID string, applicationID string, environment string) (platform.HttpResponseBusinessMoments, error) {
	return platform.HttpResponseBusinessMoments{}, nil
}

func (m GitStorageRepoMock) DeleteBusinessMoment(customerID string, applicationID string, environment string, microserviceID string, momentID string) error {
	return nil
}

func (m GitStorageRepoMock) SaveBusinessMomentEntity(customerID string, input platform.HttpInputBusinessMomentEntity) error {
	return nil
}

func (m GitStorageRepoMock) DeleteBusinessMomentEntity(customerID string, applicationID string, environment string, microserviceID string, entityID string) error {
	return nil
}

func (m GitStorageRepoMock) GetCustomers() ([]platform.Customer, error) {
	return nil, nil
}

func (m GitStorageRepoMock) SaveCustomer(customer storage.JSONCustomer) error {
	return nil
}

func (m GitStorageRepoMock) Pull() error {
	return nil
}

func (m GitStorageRepoMock) GetDirectory() string {
	return ""
}

func (m GitStorageRepoMock) CommitPathAndPush(path string, msg string) error {
	return nil
}
