package automate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	configK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func dumpConfigMap(configMap *corev1.ConfigMap) []byte {
	scheme, serializer, err := initializeSchemeAndSerializer()
	if err != nil {
		panic(err.Error())
	}

	setConfigMapGVK(scheme, configMap)

	var buf bytes.Buffer

	_ = serializer.Encode(configMap, &buf)
	return buf.Bytes()
}

func getDolittleConfigMaps(ctx context.Context, client kubernetes.Interface, namespace string) ([]corev1.ConfigMap, error) {
	results := make([]corev1.ConfigMap, 0)
	configmaps, err := client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return results, err
	}

	for _, configMap := range configmaps.Items {
		if !strings.HasSuffix(configMap.GetName(), "-dolittle") {
			continue
		}
		results = append(results, configMap)
	}
	return results, nil
}

func getOneDolittleConfigMap(ctx context.Context, client kubernetes.Interface, applicationID string, environment string, microserviceID string) (*corev1.ConfigMap, error) {
	namespace := fmt.Sprintf("application-%s", applicationID)
	var result *corev1.ConfigMap
	configmaps, err := client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return result, err
	}

	for _, configMap := range configmaps.Items {
		labels := configMap.GetLabels()
		annotations := configMap.GetAnnotations()

		if annotations["dolittle.io/microservice-id"] != microserviceID {
			continue
		}

		if labels["environment"] != environment {
			continue
		}

		if !strings.HasSuffix(configMap.GetName(), "-dolittle") {
			continue
		}

		return &configMap, nil
	}
	return result, errors.New("not.found")
}

// TODO Taken from K8sGetDeployments, we should maybe refactor this into shared command
// TODO func K8sGetDeployments(client kubernetes.Interface, context context.Context, namespace string) ([]v1.Deployment, error) {
func getDeployments(ctx context.Context, client kubernetes.Interface, namespace string) ([]appsv1.Deployment, error) {
	deployments, err := client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var microserviceDeployments []appsv1.Deployment
	for _, deployment := range deployments.Items {
		if _, ok := deployment.Labels["microservice"]; !ok {
			continue
		}
		if _, ok := deployment.Annotations["dolittle.io/microservice-id"]; !ok {
			continue
		}
		microserviceDeployments = append(microserviceDeployments, deployment)
	}
	return microserviceDeployments, nil
}

func convertObjectMetaToMicroservice(objectMeta metav1.Object) configK8s.Microservice {
	labels := objectMeta.GetLabels()
	annotations := objectMeta.GetAnnotations()

	microserviceID := annotations["dolittle.io/microservice-id"]
	microserviceName := labels["microservice"]
	customerTenant := configK8s.Tenant{
		Name: labels["tenant"],
		ID:   annotations["dolittle.io/tenant-id"],
	}
	k8sApplication := configK8s.Application{
		Name: labels["application"],
		ID:   annotations["dolittle.io/application-id"],
	}

	environment := labels["environment"]

	return configK8s.Microservice{
		ID:          microserviceID,
		Name:        microserviceName,
		Tenant:      customerTenant,
		Application: k8sApplication,
		Environment: environment,
		// TODO wwhen we explicitly set the kind as an annotation we can rely on it
		ResourceID: "",
	}
}

func getAllCustomerMicroservices(ctx context.Context, client kubernetes.Interface) ([]configK8s.Microservice, error) {
	microservices := make([]configK8s.Microservice, 0)
	deployments := make([]appsv1.Deployment, 0)
	namespaces := getNamespaces(ctx, client)
	for _, namespace := range namespaces {
		if !isApplicationNamespace(namespace) {
			continue
		}
		specific, err := getDeployments(ctx, client, namespace.Name)
		if err != nil {
			return microservices, err
		}
		deployments = append(deployments, specific...)
	}

	for _, deployment := range deployments {
		microservice := convertObjectMetaToMicroservice(deployment.GetObjectMeta())
		microservices = append(microservices, microservice)
	}
	return microservices, nil
}

func getNamespaces(ctx context.Context, client kubernetes.Interface) []corev1.Namespace {
	namespacesList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	return namespacesList.Items
}

func isApplicationNamespace(namespace corev1.Namespace) bool {
	if !strings.HasPrefix(namespace.GetName(), "application-") {
		return false
	}
	if _, hasAnnotation := namespace.Annotations["dolittle.io/tenant-id"]; !hasAnnotation {
		return false
	}
	if _, hasAnnotation := namespace.Annotations["dolittle.io/application-id"]; !hasAnnotation {
		return false
	}
	if _, hasLabel := namespace.Labels["tenant"]; !hasLabel {
		return false
	}
	if _, hasLabel := namespace.Labels["application"]; !hasLabel {
		return false
	}

	return true
}
