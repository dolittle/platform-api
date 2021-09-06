package microservice

// configmap needs to include WH_AUTHORIZATION
// hook into the deployment
// Basic: XXX
// Bearer: XXX
// foobar
import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/businessmomentsadaptor"
	networkingv1 "k8s.io/api/networking/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

type businessMomentsAdaptorRepo struct {
	k8sClient *kubernetes.Clientset
}

func NewBusinessMomentsAdaptorRepo(k8sClient *kubernetes.Clientset) businessMomentsAdaptorRepo {
	return businessMomentsAdaptorRepo{
		k8sClient: k8sClient,
	}
}

func (r businessMomentsAdaptorRepo) Create(namespace string, tenant k8s.Tenant, application k8s.Application, applicationIngress k8s.Ingress, input platform.HttpInputBusinessMomentAdaptorInfo) error {
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
	configBusinessMoments := businessmomentsadaptor.NewBusinessMomentsConfigmap(microservice)
	ingress.Spec.TLS = k8s.AddIngressTLS([]string{host}, secretName)
	ingress.Spec.Rules = append(ingress.Spec.Rules, k8s.AddIngressRule(host, ingressRules))

	token := ""

	connectorBytes, _ := json.Marshal(input.Extra.Connector)
	var whatKind platform.HttpInputMicroserviceKind
	json.Unmarshal(connectorBytes, &whatKind)
	switch whatKind.Kind {
	case "webhook":
		var connector platform.HttpInputBusinessMomentAdaptorConnectorWebhook
		json.Unmarshal(connectorBytes, &connector)

		//connectorConfigBytes, _ := json.Marshal(connector.Config.Config)
		//json.Unmarshal(connectorConfigBytes, &whatKind)
		// Super ugly
		switch connector.Config.Kind {
		case "basic":
			var connectorCredentialsConfig platform.HttpInputBusinessMomentAdaptorConnectorWebhookConfigBasic
			connectorCredentialsConfigBytes, _ := json.Marshal(connector.Config.Config)
			json.Unmarshal(connectorCredentialsConfigBytes, &connectorCredentialsConfig)
			token = fmt.Sprintf("Basic %s",
				basicAuth(connectorCredentialsConfig.Username, connectorCredentialsConfig.Password),
			)
		case "bearer":
			var connectorCredentialsConfig platform.HttpInputBusinessMomentAdaptorConnectorWebhookConfigBearer
			connectorCredentialsConfigBytes, _ := json.Marshal(connector.Config)
			json.Unmarshal(connectorCredentialsConfigBytes, &connectorCredentialsConfig)

			token = fmt.Sprintf("Bearer %s", connectorCredentialsConfig.Token)
		default:
			return errors.New("Not supported basic / bearer")
		}

	case "rest":
		return errors.New("TODO kind: rest")
	default:
		return errors.New("Not supported webhook / rest")
	}

	configEnvVariables.Data = map[string]string{
		"WH_AUTHORIZATION": token,
	}

	service.Spec.Ports[0].TargetPort = intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: 3008,
	}
	deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = 3008

	// Assuming the namespace exists
	var err error
	client := r.k8sClient
	ctx := context.TODO()

	// ConfigMaps
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
	_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, configBusinessMoments, metaV1.CreateOptions{})
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("config business moments") }) != nil {
		return err
	}
	// Secrets
	_, err = client.CoreV1().Secrets(namespace).Create(ctx, configSecrets, metaV1.CreateOptions{})
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("config secrets") }) != nil {
		return err
	}

	// Ingress
	_, err = client.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metaV1.CreateOptions{})
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("ingress") }) != nil {
		return err
	}

	// Service
	_, err = client.CoreV1().Services(namespace).Create(ctx, service, metaV1.CreateOptions{})
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("service") }) != nil {
		return err
	}

	// NetworkPolicy
	_, err = client.NetworkingV1().NetworkPolicies(namespace).Create(ctx, networkPolicy, metaV1.CreateOptions{})
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("network policy") }) != nil {
		return err
	}

	_, err = client.AppsV1().Deployments(namespace).Create(ctx, deployment, metaV1.CreateOptions{})
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("deployment") }) != nil {
		return err
	}

	return nil
}

func (r businessMomentsAdaptorRepo) Delete(namespace string, microserviceID string) error {
	client := r.k8sClient
	ctx := context.TODO()
	// Not possible to filter based on annotations
	listOpts := metaV1.ListOptions{}
	deployment, err := k8sGetDeployment(r.k8sClient, ctx, namespace, microserviceID)

	// Stop deployment
	err = k8sStopDeployment(r.k8sClient, ctx, namespace, &deployment)
	if err != nil {
		return err
	}

	// Selector information for microservice, based on labels
	listOpts = metaV1.ListOptions{
		LabelSelector: labels.FormatLabels(deployment.GetObjectMeta().GetLabels()),
	}

	// Remove configmaps
	configs, _ := client.CoreV1().ConfigMaps(namespace).List(ctx, listOpts)

	for _, config := range configs.Items {
		err = client.CoreV1().ConfigMaps(namespace).Delete(ctx, config.Name, metaV1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("todo")
		}
	}

	// Remove secrets
	secrets, _ := client.CoreV1().Secrets(namespace).List(ctx, listOpts)
	for _, secret := range secrets.Items {
		err = client.CoreV1().Secrets(namespace).Delete(ctx, secret.Name, metaV1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("todo")
		}
	}

	// Remove Ingress
	ingresses, _ := client.NetworkingV1().Ingresses(namespace).List(ctx, listOpts)
	for _, ingress := range ingresses.Items {
		err = client.NetworkingV1().Ingresses(namespace).Delete(ctx, ingress.Name, metaV1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("issue")
		}
	}

	// Remove Network Policy
	policies, _ := client.NetworkingV1().NetworkPolicies(namespace).List(ctx, listOpts)
	for _, policy := range policies.Items {
		err = client.NetworkingV1().NetworkPolicies(namespace).Delete(ctx, policy.Name, metaV1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("issue")
		}
	}

	// Remove Service
	services, _ := client.CoreV1().Services(namespace).List(ctx, listOpts)
	for _, service := range services.Items {
		err = client.CoreV1().Services(namespace).Delete(ctx, service.Name, metaV1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("issue")
		}
	}

	// Remove deployment
	err = client.AppsV1().Deployments(namespace).Delete(ctx, deployment.Name, metaV1.DeleteOptions{})
	if err != nil {
		log.Fatal(err)
		return errors.New("todo")
	}

	return nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
