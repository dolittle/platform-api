package private

import (
	"context"
	"fmt"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	microserviceK8s "github.com/dolittle/platform-api/pkg/platform/microservice/k8s"
	"k8s.io/client-go/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type PrivateMicroserviceResources struct {
	Service                    *corev1.Service
	Deployment                 *appsv1.Deployment
	DolittleConfig             *corev1.ConfigMap
	ConfigFiles                *corev1.ConfigMap
	ConfigEnvironmentVariables *corev1.ConfigMap
	SecretEnvironmentVariables *corev1.Secret
	NetworkPolicy              *networkingv1.NetworkPolicy
	RbacPolicyRules            []rbacv1.PolicyRule
}

type privateRepo struct {
	k8sClient       kubernetes.Interface
	k8sDolittleRepo platformK8s.K8sRepo
	k8sRepoV2       k8s.Repo
}

func NewPrivateRepo(k8sClient kubernetes.Interface, k8sDolittleRepo platformK8s.K8sRepo, k8sRepoV2 k8s.Repo) privateRepo {
	return privateRepo{
		k8sClient:       k8sClient,
		k8sDolittleRepo: k8sDolittleRepo,
		k8sRepoV2:       k8sRepoV2,
	}
}

func (r privateRepo) Create(
	namespace string,
	tenant dolittleK8s.Tenant,
	application dolittleK8s.Application,
	customerTenants []platform.CustomerTenantInfo,
	input platform.HttpInputSimpleInfo) error {

	client := r.k8sClient
	ctx := context.TODO()
	applicationID := application.ID

	resources := NewResources(namespace, tenant, application, customerTenants, input)
	fmt.Println("made resources")
	fmt.Println(resources)
	return nil

	_, err := client.CoreV1().ConfigMaps(namespace).Create(ctx, resources.DolittleConfig, metav1.CreateOptions{})
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

// Delete is copied from SimpleRepo's Delete, except without the ingress part
func (r privateRepo) Delete(applicationID, environment, microserviceID string) error {
	ctx := context.TODO()
	namespace := platformK8s.GetApplicationNamespace(applicationID)

	deployment, err := r.k8sRepoV2.GetDeployment(namespace, environment, microserviceID)
	if err != nil {
		return err
	}

	microserviceName := deployment.Labels["Microservice"]
	policyRules := microserviceK8s.NewMicroservicePolicyRules(microserviceName, environment)

	if err = microserviceK8s.K8sStopDeployment(r.k8sClient, ctx, namespace, &deployment); err != nil {
		return err
	}

	listOpts := metav1.ListOptions{
		LabelSelector: labels.FormatLabels(deployment.GetObjectMeta().GetLabels()),
	}

	if err = microserviceK8s.K8sDeleteConfigmaps(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = microserviceK8s.K8sDeleteSecrets(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = microserviceK8s.K8sDeleteNetworkPolicies(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = microserviceK8s.K8sDeleteServices(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

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

func NewResources(
	namespace string,
	tenant dolittleK8s.Tenant,
	application dolittleK8s.Application,
	customerTenants []platform.CustomerTenantInfo,
	input platform.HttpInputSimpleInfo,
) PrivateMicroserviceResources {
	environment := input.Environment

	microserviceID := input.Dolittle.MicroserviceID
	microserviceName := input.Name
	headImage := input.Extra.Headimage
	runtimeImage := input.Extra.Runtimeimage

	microservice := dolittleK8s.Microservice{
		ID:          microserviceID,
		Name:        microserviceName,
		Tenant:      tenant,
		Application: application,
		Environment: environment,
		Kind:        input.Kind,
	}

	// TODO if runtimeImage = none, do we need dolittleConfig? might not be any harm in keeping it around
	dolittleConfig := dolittleK8s.NewMicroserviceConfigmap(microservice, customerTenants)
	deployment := dolittleK8s.NewDeployment(microservice, headImage, runtimeImage)
	service := dolittleK8s.NewService(microservice)

	networkPolicy := dolittleK8s.NewNetworkPolicy(microservice)
	configEnvVariables := dolittleK8s.NewEnvVariablesConfigmap(microservice)
	configFiles := dolittleK8s.NewConfigFilesConfigmap(microservice)
	secretEnvVariables := dolittleK8s.NewEnvVariablesSecret(microservice)

	// Return policyRules for use with "developer"
	policyRules := microserviceK8s.NewMicroservicePolicyRules(microservice.Name, environment)

	// Ingress section
	// ingresses := customertenant.CreateIngresses(isProduction, customerTenants, microservice, service.Name, input.Extra.Ingress)

	return PrivateMicroserviceResources{
		Service:                    service,
		ConfigFiles:                configFiles,
		ConfigEnvironmentVariables: configEnvVariables,
		SecretEnvironmentVariables: secretEnvVariables,
		Deployment:                 deployment,
		DolittleConfig:             dolittleConfig,
		NetworkPolicy:              networkPolicy,
		RbacPolicyRules:            policyRules,
	}
}
