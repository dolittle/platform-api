package simple

import (
	"context"
	"fmt"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	. "github.com/dolittle-entropy/platform-api/pkg/platform/microservice/k8s"
	networkingv1 "k8s.io/api/networking/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type simpleRepo struct {
	k8sClient kubernetes.Interface
}

func NewSimpleRepo(k8sClient kubernetes.Interface) Repo {
	return &simpleRepo{
		k8sClient,
	}
}

func (r *simpleRepo) Create(namespace string, customer k8s.Tenant, application k8s.Application, ingress k8s.Ingress, tenant platform.TenantId, input platform.HttpInputSimpleInfo) error {
	// TODO not sure where this comes from really, assume dynamic
	microservice := k8s.Microservice{
		ID:          input.Dolittle.MicroserviceID,
		Name:        input.Name,
		Tenant:      customer,
		Application: application,
		Environment: input.Environment,
		ResourceID:  string(tenant),
		Kind:        platform.MicroserviceKindSimple,
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

	microserviceConfigmap := k8s.NewMicroserviceConfigmap(microservice, string(tenant))
	deployment := k8s.NewDeployment(microservice, input.Extra.Headimage, input.Extra.Runtimeimage)
	service := k8s.NewService(microservice)
	ingressResource := k8s.NewIngress(microservice)
	networkPolicy := k8s.NewNetworkPolicy(microservice)
	configEnvVariables := k8s.NewEnvVariablesConfigmap(microservice)
	configFiles := k8s.NewConfigFilesConfigmap(microservice)
	configSecrets := k8s.NewEnvVariablesSecret(microservice)

	ingressResource.Spec.TLS = k8s.AddIngressTLS([]string{ingress.Host}, ingress.SecretName)
	ingressResource.Spec.Rules = append(ingressResource.Spec.Rules, k8s.AddIngressRule(ingress.Host, ingressRules))

	// Assuming the namespace exists
	var err error
	client := r.k8sClient
	ctx := context.TODO()

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, microserviceConfigmap, metaV1.CreateOptions{})
	if K8sHandleResourceCreationError(err, func() { K8sPrintAlreadyExists("microservice config map") }) != nil {
		return err
	}

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, configEnvVariables, metaV1.CreateOptions{})
	if K8sHandleResourceCreationError(err, func() { K8sPrintAlreadyExists("config env variables") }) != nil {
		return err
	}

	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, configFiles, metaV1.CreateOptions{})
	if K8sHandleResourceCreationError(err, func() { K8sPrintAlreadyExists("config files") }) != nil {
		return err
	}

	_, err = client.CoreV1().Secrets(namespace).Create(ctx, configSecrets, metaV1.CreateOptions{})
	if K8sHandleResourceCreationError(err, func() { K8sPrintAlreadyExists("config secrets") }) != nil {
		return err
	}

	_, err = client.NetworkingV1().Ingresses(namespace).Create(ctx, ingressResource, metaV1.CreateOptions{})
	if K8sHandleResourceCreationError(err, func() { K8sPrintAlreadyExists("ingress") }) != nil {
		return err
	}

	_, err = client.NetworkingV1().NetworkPolicies(namespace).Create(ctx, networkPolicy, metaV1.CreateOptions{})
	if K8sHandleResourceCreationError(err, func() { K8sPrintAlreadyExists("network policy") }) != nil {
		return err
	}

	_, err = client.CoreV1().Services(namespace).Create(ctx, service, metaV1.CreateOptions{})
	if K8sHandleResourceCreationError(err, func() { K8sPrintAlreadyExists("service") }) != nil {
		return err
	}

	_, err = client.AppsV1().Deployments(namespace).Create(ctx, deployment, metaV1.CreateOptions{})
	if K8sHandleResourceCreationError(err, func() { K8sPrintAlreadyExists("deployment") }) != nil {
		return err
	}

	return nil
}

func (r *simpleRepo) Delete(namespace, microserviceID string) error {
	ctx := context.TODO()

	deployment, err := K8sGetDeployment(r.k8sClient, ctx, namespace, microserviceID)
	if err != nil {
		return err
	}

	if err = K8sStopDeployment(r.k8sClient, ctx, namespace, &deployment); err != nil {
		return err
	}

	// Selector information for microservice, based on labels
	listOpts := metaV1.ListOptions{
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
