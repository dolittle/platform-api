package k8s

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type repo struct {
	client     kubernetes.Interface
	logContext logrus.FieldLogger
}

type RepoIngress interface {
	GetIngresses(namespace string) ([]networkingv1.Ingress, error)
	GetIngressesWithOptions(namespace string, opts metav1.ListOptions) ([]networkingv1.Ingress, error)
	GetIngressesByEnvironmentWithMicoservice(namespace string, environment string) ([]networkingv1.Ingress, error)
}

type RepoNamspace interface {
	GetNamespaces() ([]corev1.Namespace, error)
	GetNamespacesWithOptions(opts metav1.ListOptions) ([]corev1.Namespace, error)
	GetNamespacesWithApplication() ([]corev1.Namespace, error)
}

type RepoDeployment interface {
	GetDeployments(namespace string) ([]appsv1.Deployment, error)
	GetDeploymentsWithOptions(namespace string, opts metav1.ListOptions) ([]appsv1.Deployment, error)
	GetDeploymentsWithMicroservice(namespace string) ([]appsv1.Deployment, error)
}

type Repo interface {
	RepoIngress
	RepoNamspace
	RepoDeployment
}

func NewRepo(client kubernetes.Interface, logContext logrus.FieldLogger) Repo {
	return repo{
		client:     client,
		logContext: logContext,
	}
}

func (r repo) GetIngressesByEnvironmentWithMicoservice(namespace string, environment string) ([]networkingv1.Ingress, error) {
	opts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("tenant,environment=%s,microservice", environment),
	}
	return r.GetIngressesWithOptions(namespace, opts)
}

func (r repo) GetIngressesWithOptions(namespace string, opts metav1.ListOptions) ([]networkingv1.Ingress, error) {
	ctx := context.TODO()
	items, err := r.client.NetworkingV1().Ingresses(namespace).List(ctx, opts)
	return items.Items, err
}

func (r repo) GetIngresses(namespace string) ([]networkingv1.Ingress, error) {
	opts := metav1.ListOptions{}
	return r.GetIngressesWithOptions(namespace, opts)
}

func (r repo) GetNamespacesWithApplication() ([]corev1.Namespace, error) {
	opts := metav1.ListOptions{
		LabelSelector: "tenant,application",
	}
	return r.GetNamespacesWithOptions(opts)
}
func (r repo) GetNamespacesWithOptions(opts metav1.ListOptions) ([]corev1.Namespace, error) {
	ctx := context.TODO()
	items, err := r.client.CoreV1().Namespaces().List(ctx, opts)
	return items.Items, err
}

func (r repo) GetNamespaces() ([]corev1.Namespace, error) {
	opts := metav1.ListOptions{}
	return r.GetNamespacesWithOptions(opts)
}

func (r repo) GetDeployments(namespace string) ([]appsv1.Deployment, error) {
	opts := metav1.ListOptions{}
	return r.GetDeploymentsWithOptions(namespace, opts)
}

func (r repo) GetDeploymentsWithOptions(namespace string, opts metav1.ListOptions) ([]appsv1.Deployment, error) {
	ctx := context.TODO()
	items, err := r.client.AppsV1().Deployments(namespace).List(ctx, opts)
	return items.Items, err

}

func (r repo) GetDeploymentsWithMicroservice(namespace string) ([]appsv1.Deployment, error) {
	opts := metav1.ListOptions{
		LabelSelector: "tenant,application,environment,microservice",
	}
	// TODO we could do extra filtering in here to confirm things have correct annotations?
	// Instead of in the code
	// if !IsApplicationNamespace(namespace) {
	return r.GetDeploymentsWithOptions(namespace, opts)
}
