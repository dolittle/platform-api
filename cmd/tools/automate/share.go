package automate

import (
	"context"
	"strings"

	configK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

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
