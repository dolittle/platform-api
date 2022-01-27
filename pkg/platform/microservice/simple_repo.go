package microservice

import (
	"context"
	"fmt"
	"strings"

	"github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/automate"
	. "github.com/dolittle/platform-api/pkg/platform/microservice/k8s"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type simpleRepo struct {
	k8sClient kubernetes.Interface
	kind      platform.MicroserviceKind
}

func NewSimpleRepo(k8sClient kubernetes.Interface) simpleRepo {
	return simpleRepo{
		k8sClient,
		platform.MicroserviceKindSimple,
	}
}

func (r simpleRepo) Create(namespace string, tenant k8s.Tenant, application k8s.Application, applicationIngress k8s.Ingress, input platform.HttpInputSimpleInfo) error {
	environment := input.Environment
	host := applicationIngress.Host
	secretName := applicationIngress.SecretName

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
		ResourceID:  TodoCustomersTenantID,
		Kind:        r.kind,
	}

	ingressServiceName := strings.ToLower(fmt.Sprintf("%s-%s", microservice.Environment, microservice.Name))
	ingressRules := []k8s.SimpleIngressRule{
		{
			Path:            input.Extra.Ingress.Path,
			PathType:        networkingv1.PathType(input.Extra.Ingress.Pathtype),
			ServiceName:     ingressServiceName,
			ServicePortName: "http",
		},
	}

	microserviceConfigmap := k8s.NewMicroserviceConfigmap(microservice, TodoCustomersTenantID)
	deployment := k8s.NewDeployment(microservice, headImage, runtimeImage)
	service := k8s.NewService(microservice)
	ingress := k8s.NewIngress(microservice)
	networkPolicy := k8s.NewNetworkPolicy(microservice)
	configEnvVariables := k8s.NewEnvVariablesConfigmap(microservice)
	configFiles := k8s.NewConfigFilesConfigmap(microservice)
	configSecrets := k8s.NewEnvVariablesSecret(microservice)

	ingress.Spec.TLS = k8s.AddIngressTLS([]string{host}, secretName)
	ingress.Spec.Rules = append(ingress.Spec.Rules, k8s.AddIngressRule(host, ingressRules))

	// Assuming the namespace exists
	var err error
	client := r.k8sClient
	ctx := context.TODO()

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, microserviceConfigmap, metav1.CreateOptions{})
	if K8sHandleResourceCreationError(err, func() { K8sPrintAlreadyExists("microservice config map") }) != nil {
		return err
	}

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, configEnvVariables, metav1.CreateOptions{})
	if K8sHandleResourceCreationError(err, func() { K8sPrintAlreadyExists("config env variables") }) != nil {
		return err
	}

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, configFiles, metav1.CreateOptions{})
	if K8sHandleResourceCreationError(err, func() { K8sPrintAlreadyExists("config files") }) != nil {
		return err
	}

	_, err = client.CoreV1().Secrets(namespace).Create(ctx, configSecrets, metav1.CreateOptions{})
	if K8sHandleResourceCreationError(err, func() { K8sPrintAlreadyExists("config secrets") }) != nil {
		return err
	}

	_, err = client.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if K8sHandleResourceCreationError(err, func() { K8sPrintAlreadyExists("ingress") }) != nil {
		return err
	}

	_, err = client.NetworkingV1().NetworkPolicies(namespace).Create(ctx, networkPolicy, metav1.CreateOptions{})
	if K8sHandleResourceCreationError(err, func() { K8sPrintAlreadyExists("network policy") }) != nil {
		return err
	}

	_, err = client.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if K8sHandleResourceCreationError(err, func() { K8sPrintAlreadyExists("service") }) != nil {
		return err
	}

	_, err = client.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if K8sHandleResourceCreationError(err, func() { K8sPrintAlreadyExists("deployment") }) != nil {
		return err
	}

	return nil
}

func (r simpleRepo) Delete(applicationID, environment, microserviceID string) error {
	ctx := context.TODO()

	deployment, err := automate.GetDeployment(ctx, r.k8sClient, applicationID, environment, microserviceID)
	if err != nil {
		return err
	}

	namespace := fmt.Sprintf("application-%s", applicationID)
	if err = K8sStopDeployment(r.k8sClient, ctx, namespace, &deployment); err != nil {
		return err
	}

	// Selector information for microservice, based on labels
	listOpts := metav1.ListOptions{
		LabelSelector: labels.FormatLabels(deployment.GetObjectMeta().GetLabels()),
	}

	if err = K8sDeleteConfigmaps(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = K8sDeleteSecrets(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = K8sDeleteIngresses(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = K8sDeleteNetworkPolicies(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = K8sDeleteServices(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = K8sDeleteDeployment(r.k8sClient, ctx, namespace, &deployment); err != nil {
		return err
	}
	return nil
}
