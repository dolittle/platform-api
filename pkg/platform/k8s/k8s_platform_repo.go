package k8s

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

type K8sPlatformRepo interface {
	CanModifyApplicationWithResponse(w http.ResponseWriter, customerID string, applicationID string, userID string) bool
	GetApplication(applicationID string) (platform.Application, error)
	GetMicroservices(applicationID string) ([]platform.MicroserviceInfo, error)
	GetPodStatus(applicationID string, environment string, microserviceID string) (platform.PodData, error)
	GetLogs(applicationID string, containerName string, podName string) (string, error)
	GetMicroserviceDNS(applicationID string, microserviceID string) (string, error)
	GetConfigMap(applicationID string, name string) (*corev1.ConfigMap, error)
	GetSecret(logContext logrus.FieldLogger, applicationID string, name string) (*corev1.Secret, error)
	GetServiceAccount(logContext logrus.FieldLogger, applicationID string, name string) (*corev1.ServiceAccount, error)
	CreateServiceAccountFromResource(logContext logrus.FieldLogger, resource *corev1.ServiceAccount) (*corev1.ServiceAccount, error)
	CreateServiceAccount(logger logrus.FieldLogger, customerID string, customerName string, applicationID string, applicationName string, serviceAccountName string) (*corev1.ServiceAccount, error)
	AddServiceAccountToRoleBinding(logger logrus.FieldLogger, applicationID string, roleBinding string, serviceAccount string) (*rbacv1.RoleBinding, error)
	CreateRoleBinding(logger logrus.FieldLogger, customerID, customerName, applicationID, applicationName, roleBinding, role string) (*rbacv1.RoleBinding, error)
	CanModifyApplication(customerID string, applicationID string, userID string) (bool, error)
	CanModifyApplicationWithResourceAttributes(customerID string, applicationID string, userID string, attribute authv1.ResourceAttributes) (bool, error)
	GetRestConfig() *rest.Config
	GetIngressURLsWithCustomerTenantID(ingresses []networkingv1.Ingress, microserviceID string) ([]platform.IngressURLWithCustomerTenantID, error)
	GetIngressHTTPIngressPath(ingresses []networkingv1.Ingress, microserviceID string) ([]networkingv1.HTTPIngressPath, error)
	GetApplications(customerID string) ([]platform.ShortInfo, error)
	GetMicroserviceName(applicationID string, environment string, microserviceID string) (string, error)
	RestartMicroservice(applicationID string, environment string, microserviceID string) error
	WriteConfigMap(configMap *corev1.ConfigMap) (*corev1.ConfigMap, error)
	WriteSecret(secret *corev1.Secret) (*corev1.Secret, error)
	GetUserSpecificSubjectRulesReviewStatus(applicationID string, groupID string, userID string) (authv1.SubjectRulesReviewStatus, error)
	RemovePolicyRule(roleName string, applicationID string, newRule rbacv1.PolicyRule) error
	AddPolicyRule(roleName string, applicationID string, newRule rbacv1.PolicyRule) error
	AddServiceAccount(serviceAccount string, roleBinding string, customerID string, customerName string, applicationID string, applicationName string) error
}
