package k8s

import (
	"github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/customertenant"
	microserviceK8s "github.com/dolittle/platform-api/pkg/platform/microservice/k8s"
	corev1 "k8s.io/api/core/v1"
)

func NewResources(
	isProduction bool,
	namespace string,
	tenant k8s.Tenant,
	application k8s.Application,
	customerTenants []platform.CustomerTenantInfo,
	input platform.HttpInputSimpleInfo,
) SimpleMicroserviceResources {
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

	var dolittleConfig *corev1.ConfigMap
	switch runtimeImage {
	case "dolittle/runtime:6.1.0":
		dolittleConfig = k8s.NewMicroserviceConfigmapV6_1_0(microservice, customerTenants)
	case "none":
		fallthrough
	default:
		dolittleConfig = k8s.NewMicroserviceConfigmap(microservice, customerTenants)
	}

	deployment := k8s.NewDeployment(microservice, headImage, runtimeImage)
	service := k8s.NewService(microservice)

	networkPolicy := k8s.NewNetworkPolicy(microservice)
	configEnvVariables := k8s.NewEnvVariablesConfigmap(microservice)
	configFiles := k8s.NewConfigFilesConfigmap(microservice)
	secretEnvVariables := k8s.NewEnvVariablesSecret(microservice)

	// Return policyRules for use with "developer"
	policyRules := microserviceK8s.NewMicroservicePolicyRules(microservice.Name, environment)

	// Ingress section
	ingresses := customertenant.CreateIngresses(isProduction, customerTenants, microservice, service.Name, input.Extra.Ingress)

	return SimpleMicroserviceResources{
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
