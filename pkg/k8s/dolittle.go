package k8s

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
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
	GetIngressesByEnvironmentWithMicoservices(namespace string, environment string) ([]networkingv1.Ingress, error)
}

type RepoNamspace interface {
	GetNamespaces() ([]corev1.Namespace, error)
	GetNamespacesWithOptions(opts metav1.ListOptions) ([]corev1.Namespace, error)
	GetNamespacesWithApplication() ([]corev1.Namespace, error)
}

type Repo interface {
	RepoIngress
	RepoNamspace
}

func NewRepo(client kubernetes.Interface, logContext logrus.FieldLogger) Repo {
	return repo{
		client:     client,
		logContext: logContext,
	}
}

func (r repo) GetIngressesByEnvironmentWithMicoservices(namespace string, environment string) ([]networkingv1.Ingress, error) {
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
