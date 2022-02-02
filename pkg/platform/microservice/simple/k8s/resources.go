package k8s

import (
	"fmt"
	"strings"

	"github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	microserviceK8s "github.com/dolittle/platform-api/pkg/platform/microservice/k8s"

	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// TODO Refactor to accept the customerTenantId https://github.com/dolittle/platform-api/pull/65
func NewResources(
	platformEnvironment string,
	namespace string,
	tenant k8s.Tenant,
	application k8s.Application,
	customerTenants []platform.CustomerTenantInfo,
	subjects []rbacv1.Subject, // TODO this might not be needed
	input platform.HttpInputSimpleInfo,
) MicroserviceResources {
	environment := input.Environment

	microserviceID := input.Dolittle.MicroserviceID
	microserviceName := input.Name
	headImage := input.Extra.Headimage
	runtimeImage := input.Extra.Runtimeimage

	microservice := k8s.Microservice{
		ID:          microserviceID,
		Name:        microserviceName,
		Tenant:      tenant,
		Application: application,
		Environment: environment,
		Kind:        input.Kind,
	}

	// TODO if runtimeImage = none, do we need dolittleConfig? might not be any harm in keeping it around
	// TODO if we do not need the runtime, do we need a customerTenantID? (simple answer is keep it)
	dolittleConfig := k8s.NewMicroserviceConfigmap(microservice, customerTenants)
	deployment := k8s.NewDeployment(microservice, headImage, runtimeImage)
	service := k8s.NewService(microservice)

	networkPolicy := k8s.NewNetworkPolicy(microservice)
	configEnvVariables := k8s.NewEnvVariablesConfigmap(microservice)
	configFiles := k8s.NewConfigFilesConfigmap(microservice)
	secretEnvVariables := k8s.NewEnvVariablesSecret(microservice)

	// TODO This is wrong
	// Should return policyRules for use with "developer"
	//role, roleBinding := microserviceK8s.NewMicroserviceRbac(microservice.Name, microservice.ID, string(microservice.Kind), tenant, application, environment, subjects)
	policyRules := microserviceK8s.NewMicroservicePolicyRoles(microservice.Name, environment)

	// Ingress section
	ingressServiceName := strings.ToLower(fmt.Sprintf("%s-%s", microservice.Environment, microservice.Name))
	ingressRules := []k8s.SimpleIngressRule{
		{
			Path:            input.Extra.Ingress.Path,
			PathType:        networkingv1.PathType(input.Extra.Ingress.Pathtype),
			ServiceName:     ingressServiceName,
			ServicePortName: "http",
		},
	}

	ingresses := make([]*networkingv1.Ingress, 0)
	for _, customerTenant := range customerTenants {
		// TODO needs a name / currently using indexID
		ingress := k8s.NewMicroserviceIngressWithEmptyRules(platformEnvironment, microservice)
		// TODO This could be the customerTenantID
		// TODO this could be hashed {env}-{hash}

		newName := fmt.Sprintf("%s-%s", ingress.ObjectMeta.Name, customerTenant.CustomerTenantID[0:7])

		ingress.ObjectMeta.Name = newName
		ingress = k8s.AddCustomerTenantIDToIngress(customerTenant.CustomerTenantID, ingress)
		ingress.Spec.TLS = k8s.AddIngressTLS([]string{customerTenant.Ingress.Host}, customerTenant.Ingress.SecretName)
		ingress.Spec.Rules = append(ingress.Spec.Rules, k8s.AddIngressRule(customerTenant.Ingress.Host, ingressRules))

		ingresses = append(ingresses, ingress)
	}

	return MicroserviceResources{
		Service:                    service,
		ConfigFiles:                configFiles,
		ConfigEnvironmentVariables: configEnvVariables,
		SecretEnvironmentVariables: secretEnvVariables,
		Deployment:                 deployment,
		DolittleConfig:             dolittleConfig,
		NetworkPolicy:              networkPolicy,
		Ingresses:                  ingresses,
		RbacPolicyRules:            policyRules,
	}
}
