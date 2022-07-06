package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/containerregistry"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	authv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/client-go/rest"
)

func createRequest(path, xSharedSecret string) *http.Request {
	request, _ := http.NewRequest("GET", path, nil)
	if xSharedSecret == "" {
		request.Header.Del("x-shared-secret")
	} else {
		request.Header.Set("x-shared-secret", xSharedSecret)
	}
	request.Header.Set("Tenant-ID", "123")
	request.Header.Set("User-ID", "666")
	return request
}

var _ = Describe("Platform API", func() {
	var gitRepo GitStorageRepoMock
	var containerRegistryRepo *ContainerRegistryMock
	var server *httptest.Server
	var crImagesPath string
	var crTagsPath string
	var c http.Client

	BeforeEach(func() {
		viper.Set("tools.server.gitRepo.branch", "foo")
		viper.Set("tools.server.gitRepo.directory", "/tmp/dolittle_operation")
		viper.Set("tools.server.gitRepo.url", "git@github.com:dolittle-platform/Operations")
		viper.Set("tools.server.gitRepo.sshKey", "does/not/exist")
		viper.Set("tools.server.kubernetes.externalClusterHost", "external-host")
		viper.Set("tools.server.secret", "johnc")

		gitRepo = GitStorageRepoMock{}
		containerRegistryRepo = &ContainerRegistryMock{}
		containerRegistryRepo.StubAndReturnImages([]string{"helloworld"})

		logContext := logrus.StandardLogger()
		k8sClient, k8sConfig := platformK8s.InitKubernetesClient()

		k8sPlatformRepoMock := &K8sPlatformRepoMock{}
		k8sPlatformRepoMock.StubGetSecretAndReturn(&corev1.Secret{})
		k8sRepoV2 := k8s.NewRepo(k8sClient, logContext.WithField("context", "k8s-repo-v2"))

		srv := NewServer(logContext, gitRepo, k8sClient, k8sPlatformRepoMock, k8sRepoV2, k8sConfig, containerRegistryRepo)
		server = httptest.NewServer(srv.Handler)
		crImagesPath = fmt.Sprintf("%s/application/12321/containerregistry/images", server.URL)
		crTagsPath = fmt.Sprintf("%s/application/12321/containerregistry/tags/helloworld", server.URL)

		c = http.Client{}
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("When fetch images from Container Registry", func() {

		It("should return 403 Forbidden when x-shared-secret header is not set", func() {
			request := createRequest(crImagesPath, "")
			response, _ := c.Do(request)

			Expect(response).ToNot(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("should rertun 403 Forbidden when x-shared-secret header is invalid", func() {
			request := createRequest(crImagesPath, "invalid header")
			response, _ := c.Do(request)

			Expect(response).ToNot(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("should include message in the response when x-shared-secret header is invalid", func() {
			request := createRequest(crImagesPath, "invalid header")
			response, _ := c.Do(request)

			r, _ := ioutil.ReadAll(response.Body)
			var jsonData map[string]interface{}
			json.Unmarshal(r, &jsonData)
			Expect(jsonData["message"]).To(Equal("Shared secret is wrong"))
		})

		It("should return container registry URL from the customer's config in our db (git storage)", func() {
			request := createRequest(crImagesPath, "johnc")
			response, _ := c.Do(request)

			Expect(response).ToNot(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			r, _ := ioutil.ReadAll(response.Body)
			var jsonData map[string]interface{}
			json.Unmarshal(r, &jsonData)
			Expect(jsonData["url"]).To(Equal("test1.azurecr.io"))
		})

		It("should return the images", func() {
			request := createRequest(crImagesPath, "johnc")
			response, _ := c.Do(request)

			Expect(response).ToNot(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			r, _ := ioutil.ReadAll(response.Body)
			var jsonData map[string]interface{}
			json.Unmarshal(r, &jsonData)
			Expect(jsonData["images"]).To(Equal([]interface{}{"helloworld"}))
		})

	})

	Describe("When fetch tags from the images", func() {
		It("should return the tags", func() {
			request := createRequest(crTagsPath, "johnc")
			response, _ := c.Do(request)

			Expect(response).ToNot(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			r, _ := ioutil.ReadAll(response.Body)
			var jsonData map[string]interface{}
			json.Unmarshal(r, &jsonData)
			Expect(jsonData["tags"]).To(Equal([]interface{}{"latest", "v1"}))
		})
	})

})

type ContainerRegistryMock struct {
	imagesResult []string
}

func (m *ContainerRegistryMock) StubAndReturnImages(result []string) {
	m.imagesResult = result
}

func (m *ContainerRegistryMock) GetImages(credentials containerregistry.ContainerRegistryCredentials) ([]string, error) {
	return m.imagesResult, nil
}

func (m *ContainerRegistryMock) GetTags(credentials containerregistry.ContainerRegistryCredentials, image string) ([]string, error) {
	return []string{"latest", "v1"}, nil
}

type K8sPlatformRepoMock struct {
	getSecretResult *corev1.Secret
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

func (m *K8sPlatformRepoMock) StubGetSecretAndReturn(secret *corev1.Secret) {
	m.getSecretResult = secret
}

func (m *K8sPlatformRepoMock) GetSecret(logContext logrus.FieldLogger, applicationID string, name string) (*corev1.Secret, error) {
	return m.getSecretResult, nil
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
	return authv1.SubjectRulesReviewStatus{}, nil
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

func (m *K8sPlatformRepoMock) AddServiceAccount(serviceAccount string, roleBinding string, customerID string, customerName string, applicationID string, applicationName string) error {
	return nil
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
