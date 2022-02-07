package k8s

import (
	"context"

	"github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/automate"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	microserviceK8s "github.com/dolittle/platform-api/pkg/platform/microservice/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/simple"
	"k8s.io/client-go/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/labels"
)

// TODO this name is used in dolittle k8s to describe runtime specific details
type MicroserviceResources struct {
	Service                    *corev1.Service
	Deployment                 *appsv1.Deployment
	DolittleConfig             *corev1.ConfigMap
	ConfigFiles                *corev1.ConfigMap
	ConfigEnvironmentVariables *corev1.ConfigMap
	SecretEnvironmentVariables *corev1.Secret
	NetworkPolicy              *networkingv1.NetworkPolicy
	Ingresses                  []*networkingv1.Ingress
	RbacPolicyRules            []rbacv1.PolicyRule
}

type k8sRepo struct {
	k8sClient       kubernetes.Interface
	k8sDolittleRepo platformK8s.K8sRepo
	kind            platform.MicroserviceKind
	isProduction    bool
}

func NewSimpleRepo(k8sClient kubernetes.Interface, k8sDolittleRepo platformK8s.K8sRepo, isProduction bool) simple.Repo {
	return k8sRepo{
		k8sClient:       k8sClient,
		k8sDolittleRepo: k8sDolittleRepo,
		kind:            platform.MicroserviceKindSimple,
		isProduction:    isProduction,
	}
}

func (r k8sRepo) Create(namespace string, tenant k8s.Tenant, application k8s.Application, customerTenants []platform.CustomerTenantInfo, input platform.HttpInputSimpleInfo) error {
	// Can we use dryRun?
	var err error

	client := r.k8sClient
	ctx := context.TODO()
	applicationID := application.ID

	// TODO we can remove subjects
	resources := NewResources(r.isProduction, namespace, tenant, application, customerTenants, make([]rbacv1.Subject, 0), input)

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, resources.DolittleConfig, metav1.CreateOptions{})
	if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("microservice config map") }) != nil {
		return err
	}

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, resources.ConfigEnvironmentVariables, metav1.CreateOptions{})
	if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("config env variables") }) != nil {
		return err
	}

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, resources.ConfigFiles, metav1.CreateOptions{})
	if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("config files") }) != nil {
		return err
	}

	_, err = client.CoreV1().Secrets(namespace).Create(ctx, resources.SecretEnvironmentVariables, metav1.CreateOptions{})
	if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("config secrets") }) != nil {
		return err
	}

	for _, ingress := range resources.Ingresses {
		_, err = client.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
		if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("ingress") }) != nil {
			return err
		}
	}

	_, err = client.NetworkingV1().NetworkPolicies(namespace).Create(ctx, resources.NetworkPolicy, metav1.CreateOptions{})
	if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("network policy") }) != nil {
		return err
	}

	_, err = client.CoreV1().Services(namespace).Create(ctx, resources.Service, metav1.CreateOptions{})
	if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("service") }) != nil {
		return err
	}

	_, err = client.AppsV1().Deployments(namespace).Create(ctx, resources.Deployment, metav1.CreateOptions{})
	if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("deployment") }) != nil {
		return err
	}

	// Update developer
	for _, policyRule := range resources.RbacPolicyRules {
		err = r.k8sDolittleRepo.AddPolicyRule("developer", applicationID, policyRule)
		// Not sure this is the best error checking
		if microserviceK8s.K8sHandleResourceCreationError(err, func() { microserviceK8s.K8sPrintAlreadyExists("policy rule") }) != nil {
			return err
		}
	}

	return nil
}

func (r k8sRepo) Delete(applicationID, environment, microserviceID string) error {
	ctx := context.TODO()
	namespace := platformK8s.GetApplicationNamespace(applicationID)

	deployment, err := automate.GetDeployment(ctx, r.k8sClient, applicationID, environment, microserviceID)
	if err != nil {
		return err
	}
	// TODO can i get the name? it should be the deployment name
	// Label Environment micoservice
	microserviceName := deployment.Labels["Microservice"]
	policyRules := microserviceK8s.NewMicroservicePolicyRules(microserviceName, environment)

	if err = microserviceK8s.K8sStopDeployment(r.k8sClient, ctx, namespace, &deployment); err != nil {
		return err
	}

	// Selector information for microservice, based on labels
	listOpts := metav1.ListOptions{
		LabelSelector: labels.FormatLabels(deployment.GetObjectMeta().GetLabels()),
	}

	// TODO I wonder if the order matters
	if err = microserviceK8s.K8sDeleteConfigmaps(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = microserviceK8s.K8sDeleteSecrets(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = microserviceK8s.K8sDeleteIngresses(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = microserviceK8s.K8sDeleteNetworkPolicies(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = microserviceK8s.K8sDeleteServices(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	// Remove policy rules from developer
	// This might not be the best way if we change things and use this function for cleaning up, but it works
	for _, policyRule := range policyRules {
		err := r.k8sDolittleRepo.RemovePolicyRule("developer", applicationID, policyRule)
		if err != nil {
			return err
		}
	}

	if err = microserviceK8s.K8sDeleteDeployment(r.k8sClient, ctx, namespace, &deployment); err != nil {
		return err
	}
	return nil
}
