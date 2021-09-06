package microservice

import (
	"context"
	"fmt"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	networkingv1 "k8s.io/api/networking/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type simpleRepo struct {
	k8sClient *kubernetes.Clientset
	kind      string
}

func NewSimpleRepo(k8sClient *kubernetes.Clientset) simpleRepo {
	return simpleRepo{
		k8sClient,
		platform.MicroserviceKinds.Simple,
	}
}

func (r simpleRepo) Create(namespace string, tenant k8s.Tenant, application k8s.Application, applicationIngress k8s.Ingress, input platform.HttpInputSimpleInfo) error {
	// TODO not sure where this comes from really, assume dynamic

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
		ResourceID:  todoCustomersTenantID,
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

	microserviceConfigmap := k8s.NewMicroserviceConfigmap(microservice, todoCustomersTenantID)
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

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, microserviceConfigmap, metaV1.CreateOptions{})
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("microservice config map") }) != nil {
		return err
	}

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, configEnvVariables, metaV1.CreateOptions{})
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("config env variables") }) != nil {
		return err
	}

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, configFiles, metaV1.CreateOptions{})
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("config files") }) != nil {
		return err
	}

	_, err = client.CoreV1().Secrets(namespace).Create(ctx, configSecrets, metaV1.CreateOptions{})
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("config secrets") }) != nil {
		return err
	}

	_, err = client.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metaV1.CreateOptions{})
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("ingress") }) != nil {
		return err
	}

	_, err = client.NetworkingV1().NetworkPolicies(namespace).Create(ctx, networkPolicy, metaV1.CreateOptions{})
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("network policy") }) != nil {
		return err
	}

	_, err = client.CoreV1().Services(namespace).Create(ctx, service, metaV1.CreateOptions{})
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("service") }) != nil {
		return err
	}

	_, err = client.AppsV1().Deployments(namespace).Create(ctx, deployment, metaV1.CreateOptions{})
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("deployment") }) != nil {
		return err
	}

	return nil
}

func (r simpleRepo) Delete(namespace string, microserviceID string) error {
	ctx := context.TODO()

	deployment, err := k8sGetDeployment(r.k8sClient, ctx, namespace, microserviceID)
	if err != nil {
		return err
	}

	if err = k8sStopDeployment(r.k8sClient, ctx, namespace, &deployment); err != nil {
		return err
	}

	// Selector information for microservice, based on labels
	listOpts := metaV1.ListOptions{
		LabelSelector: labels.FormatLabels(deployment.GetObjectMeta().GetLabels()),
	}

	if err = k8sDeleteConfigmaps(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = k8sDeleteSecrets(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = k8sDeleteIngresses(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = k8sDeleteNetworkPolicies(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = k8sDeleteServices(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = k8sDeleteDeployment(r.k8sClient, ctx, namespace, &deployment); err != nil {
		return err
	}
	return nil
}
