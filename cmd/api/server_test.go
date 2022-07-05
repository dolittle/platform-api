package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
)

var _ = Describe("Platform API", func() {

	Describe("When fetch images from Container Registry", func() {

		It("should return the images", func() {
			viper.Set("tools.server.gitRepo.branch", "foo")
			viper.Set("tools.server.gitRepo.directory", "/tmp/dolittle_operation")
			viper.Set("tools.server.gitRepo.url", "git@github.com:dolittle-platform/Operations")
			viper.Set("tools.server.gitRepo.sshKey", "does/not/exist")
			viper.Set("tools.server.kubernetes.externalClusterHost", "external-host")
			viper.Set("tools.server.secret", "johnc")

			logContext := logrus.StandardLogger()
			k8sClient, k8sConfig := platformK8s.InitKubernetesClient()
			gitRepo := GitStorageRepoMock{}

			//k8sRepo := platformK8s.NewK8sRepo(k8sClient, k8sConfig, logContext.WithField("context", "k8s-repo"))
			k8sPlatformRepoMock := &K8sPlatformRepoMock{}
			k8sRepoV2 := k8s.NewRepo(k8sClient, logContext.WithField("context", "k8s-repo-v2"))

			srv := NewServer(logContext, gitRepo, k8sClient, k8sPlatformRepoMock, k8sRepoV2, k8sConfig)
			s := httptest.NewServer(srv.Handler)
			defer s.Close()

			//Expect(s.URL).To(Equal("foo"))

			c := http.Client{}
			request, _ := http.NewRequest("GET", fmt.Sprintf("%s/application/12321/containerregistry/images", s.URL), nil)
			request.Header.Set("x-shared-secret", "johnc")
			request.Header.Set("Tenant-ID", "123")
			request.Header.Set("User-ID", "666")

			response, _ := c.Do(request)

			Expect(response).ToNot(BeNil())
			r, _ := ioutil.ReadAll(response.Body)
			Expect(string(r)).To(Equal("Foo"))
			Expect(response.StatusCode).To(Equal(http.StatusOK))
		})
	})

})

type K8sPlatformRepoMock struct {
}

func (m *K8sPlatformRepoMock) GetApplication(applicationID string) (platform.Application, error) {
	return platform.Application{}, nil
}

func (m *K8sPlatformRepoMock) GetMicroservices(applicationID string) ([]platform.MicroserviceInfo, error) {
	return nil, nil
}

func (m *K8sPlatformRepoMock) GetPodStatus(applicationID string, environment string, microserviceID string) (platform.PodData, error) {
	return platform.PodData{}, nil
}

func (m *K8sPlatformRepoMock) GetLogs(applicationID string, containerName string, podName string) (string, error) {
	return "", nil
}

func (m *K8sPlatformRepoMock) GetMicroserviceDNS(applicationID string, microserviceID string) (string, error) {
	return "", nil
}

func (m *K8sPlatformRepoMock) GetConfigMap(applicationID string, name string) (*corev1.ConfigMap, error) {
	return nil, nil
}

func (m *K8sPlatformRepoMock) GetSecret(logContext logrus.FieldLogger, applicationID string, name string) (*corev1.Secret, error) {
	return nil, nil
}

func (m *K8sPlatformRepoMock) GetServiceAccount(logContext logrus.FieldLogger, applicationID string, name string) (*corev1.ServiceAccount, error) {
	return nil, nil
}

func (m *K8sPlatformRepoMock) CreateServiceAccountFromResource(logContext logrus.FieldLogger, resource *corev1.ServiceAccount) (*corev1.ServiceAccount, error) {
	return nil, nil
}

func (m *K8sPlatformRepoMock) CreateServiceAccount(logger logrus.FieldLogger, customerID string, customerName string, applicationID string, applicationName string, serviceAccountName string) (*corev1.ServiceAccount, error) {
	return nil, nil
}

func (m *K8sPlatformRepoMock) AddServiceAccountToRoleBinding(logger logrus.FieldLogger, applicationID string, roleBinding string, serviceAccount string) (*rbacv1.RoleBinding, error) {
	return nil, nil
}

func (m *K8sPlatformRepoMock) CreateRoleBinding(logger logrus.FieldLogger, customerID, customerName, applicationID, applicationName, roleBinding, role string) (*rbacv1.RoleBinding, error) {
	return nil, nil
}

func (m *K8sPlatformRepoMock) CanModifyApplication(customerID string, applicationID string, userID string) (bool, error) {
	return false, nil
}

func (m *K8sPlatformRepoMock) CanModifyApplicationWithResourceAttributes(customerID string, applicationID string, userID string, attribute authv1.ResourceAttributes) (bool, error) {
	return false, nil
}

func (m *K8sPlatformRepoMock) GetRestConfig() *rest.Config {
	return &rest.Config{}
}

func (m *K8sPlatformRepoMock) GetIngressURLsWithCustomerTenantID(ingresses []networkingv1.Ingress, microserviceID string) ([]platform.IngressURLWithCustomerTenantID, error) {
	return nil, nil
}

func (m *K8sPlatformRepoMock) GetIngressHTTPIngressPath(ingresses []networkingv1.Ingress, microserviceID string) ([]networkingv1.HTTPIngressPath, error) {
	return nil, nil
}

func (m *K8sPlatformRepoMock) GetApplications(customerID string) ([]platform.ShortInfo, error) {
	return nil, nil
}

func (m *K8sPlatformRepoMock) GetMicroserviceName(applicationID string, environment string, microserviceID string) (string, error) {
	return "", nil
}

func (m *K8sPlatformRepoMock) RestartMicroservice(applicationID string, environment string, microserviceID string) error {
	return nil
}

func (m *K8sPlatformRepoMock) WriteConfigMap(configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	return nil, nil
}

func (m *K8sPlatformRepoMock) WriteSecret(secret *corev1.Secret) (*corev1.Secret, error) {
	return nil, nil
}

func (m *K8sPlatformRepoMock) GetUserSpecificSubjectRulesReviewStatus(applicationID string, groupID string, userID string) (authv1.SubjectRulesReviewStatus, error) {
	return nil, nil
}

func (m *K8sPlatformRepoMock) RemovePolicyRule(roleName string, applicationID string, newRule rbacv1.PolicyRule) error {
	return nil
}

func (m *K8sPlatformRepoMock) AddPolicyRule(roleName string, applicationID string, newRule rbacv1.PolicyRule) error {
	return nil
}

func (m *K8sPlatformRepoMock) CanModifyApplicationWithResponse(w http.ResponseWriter, customerID string, applicationID string, userID string) bool {
	return true
}

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
