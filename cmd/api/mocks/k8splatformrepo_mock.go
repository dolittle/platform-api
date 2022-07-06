package mocks

import (
	"net/http"

	"github.com/dolittle/platform-api/pkg/platform"

	"github.com/sirupsen/logrus"
	authv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/client-go/rest"
)

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
