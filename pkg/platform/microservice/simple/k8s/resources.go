package k8s

import (
	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/customertenant"
	microserviceK8s "github.com/dolittle/platform-api/pkg/platform/microservice/k8s"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewResources(
	isProduction bool,
	namespace string,
	tenant dolittleK8s.Tenant,
	application dolittleK8s.Application,
	customerTenants []platform.CustomerTenantInfo,
	input platform.HttpInputSimpleInfo,
) MicroserviceResources {
	environment := input.Environment

	microserviceID := input.Dolittle.MicroserviceID
	microserviceName := input.Name
	runtimeImage := input.Extra.Runtimeimage

	microservice := dolittleK8s.Microservice{
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
		dolittleConfig = dolittleK8s.NewMicroserviceConfigmapV6_1_0(microservice, customerTenants)
	case "dolittle/runtime:8.0.0":
		dolittleConfig = dolittleK8s.NewMicroserviceConfigmapV8_0_0(microservice, customerTenants)
	case "none":
		fallthrough
	default:
		dolittleConfig = dolittleK8s.NewMicroserviceConfigmap(microservice, customerTenants)
	}

	if input.Extra.HeadPort == 0 {
		input.Extra.HeadPort = 80
	}

	deployment := NewDeployment(microservice, input.Extra)
	service := NewService(microservice, input.Extra)

	configEnvVariables := dolittleK8s.NewEnvVariablesConfigmap(microservice)
	configFiles := dolittleK8s.NewConfigFilesConfigmap(microservice)
	secretEnvVariables := dolittleK8s.NewEnvVariablesSecret(microservice)

	// Return policyRules for use with "developer"
	policyRules := microserviceK8s.NewMicroservicePolicyRules(microservice.Name, environment)

	var ingressResources *IngressResources
	if input.Extra.Ispublic {
		ingressResources = &IngressResources{
			NetworkPolicy: dolittleK8s.NewNetworkPolicy(microservice),
			Ingresses:     customertenant.CreateIngresses(isProduction, customerTenants, microservice, service.Name, input.Extra.Ingress),
		}
	}

	return MicroserviceResources{
		Service:                    service,
		ConfigFiles:                configFiles,
		ConfigEnvironmentVariables: configEnvVariables,
		SecretEnvironmentVariables: secretEnvVariables,
		Deployment:                 deployment,
		DolittleConfig:             dolittleConfig,
		RbacPolicyRules:            policyRules,
		IngressResources:           ingressResources,
	}
}

// NewDeployment, wrapping the base deployment and making it possible to override the ContainerPort
func NewDeployment(microservice dolittleK8s.Microservice, extra platform.HttpInputSimpleExtra) *appsv1.Deployment {
	headImage := extra.Headimage
	runtimeImage := extra.Runtimeimage
	headCommand := extra.Headcommand

	deployment := dolittleK8s.NewDeployment(microservice, headImage, runtimeImage)

	// the head container should always be the first container so we can trust in that for now
	deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = extra.HeadPort
	// the Command property is equal to Dockers ENTRYPOINT, check https://stackoverflow.com/a/66078726/5806412
	deployment.Spec.Template.Spec.Containers[0].Command = headCommand.Commands
	// the Args property is equal to Dockers CMD
	deployment.Spec.Template.Spec.Containers[0].Args = headCommand.Args
	return deployment
}

// NewService, wrapping the base deployment and making it possible to override the ContainerPort
func NewService(microservice dolittleK8s.Microservice, extra platform.HttpInputSimpleExtra) *corev1.Service {
	service := dolittleK8s.NewService(microservice)

	service.Spec.Ports[0].Port = extra.HeadPort
	service.Spec.Ports[0].TargetPort = intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: extra.HeadPort,
	}
	return service
}
