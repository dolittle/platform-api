package k8s

import (
	"github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/customertenant"
	microserviceK8s "github.com/dolittle/platform-api/pkg/platform/microservice/k8s"

	rbacv1 "k8s.io/api/rbac/v1"
)

// TODO Refactor to accept the customerTenantId https://github.com/dolittle/platform-api/pull/65
func NewResources(
	isProduction bool,
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

	// Return policyRules for use with "developer"
	policyRules := microserviceK8s.NewMicroservicePolicyRules(microservice.Name, environment)

	// Ingress section
	ingresses := customertenant.CreateIngresses(isProduction, customerTenants, microservice, service.Name, input.Extra.Ingress)

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
