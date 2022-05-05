package api

import (
	"net/http/httptest"

	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var _ = Describe("foo", func() {

	It("should do something", func() {
		viper.Set("tools.server.gitRepo.branch", "foo")
		viper.Set("tools.server.gitRepo.directory", "/tmp/dolittle_operation")
		viper.Set("tools.server.gitRepo.url", "git@github.com:dolittle-platform/Operations")
		viper.Set("tools.server.gitRepo.sshKey", "does/not/exist")
		viper.Set("tools.server.kubernetes.externalClusterHost", "external-host")

		logContext := logrus.StandardLogger()
		k8sClient, k8sConfig := platformK8s.InitKubernetesClient()
		gitRepo := GitStorageRepoMock{}

		srv := NewServer(logContext, gitRepo, k8sClient, k8sConfig)
		s := httptest.NewServer(srv.Handler)
		defer s.Close()

		Expect(s.URL).To(Equal("foo"))
	})

})

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
	return platform.TerraformCustomer{}, nil
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
