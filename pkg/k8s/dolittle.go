package k8s

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	ErrAlreadyExists = errors.New("already-exists")
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
	GetDeploymentsByEnvironmentWithMicroservice(namespace string, environment string) ([]appsv1.Deployment, error)
}

type RepoRoleBinding interface {
	AddSubjectToRoleBinding(namespace string, name string, subject rbacv1.Subject) error
	RemoveSubjectToRoleBinding(namespace string, name string, subject rbacv1.Subject) error
	GetRoleBinding(namespace string, name string) (rbacv1.RoleBinding, error)
}
type Repo interface {
	RepoIngress
	RepoNamspace
	RepoDeployment
	RepoRoleBinding
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

func (r repo) GetDeploymentsByEnvironmentWithMicroservice(namespace string, environment string) ([]appsv1.Deployment, error) {
	opts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("tenant,application,environment=%s,microservice", environment),
	}
	// TODO we could do extra filtering in here to confirm things have correct annotations?
	// Instead of in the code
	// if !IsApplicationNamespace(namespace) {
	return r.GetDeploymentsWithOptions(namespace, opts)
}

func (r repo) AddSubjectToRoleBinding(namespace string, name string, subject rbacv1.Subject) error {
	roleBinding, err := r.GetRoleBinding(namespace, name)
	if err != nil {
		return err
	}

	for _, current := range roleBinding.Subjects {
		if current.Kind == subject.Kind && current.Name == subject.Name {
			return ErrAlreadyExists
		}
	}

	roleBinding.Subjects = append(roleBinding.Subjects, subject)

	ctx := context.TODO()
	_, err = r.client.RbacV1().RoleBindings(namespace).Update(ctx, &roleBinding, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (r repo) RemoveSubjectToRoleBinding(namespace string, name string, subject rbacv1.Subject) error {
	roleBinding, err := r.GetRoleBinding(namespace, name)
	if err != nil {
		return err
	}

	updateSubjects := funk.Filter(roleBinding.Subjects, func(current rbacv1.Subject) bool {
		if current.Kind != subject.Kind {
			return true
		}

		return current.Name != subject.Name
	}).([]rbacv1.Subject)

	roleBinding.Subjects = updateSubjects

	ctx := context.TODO()
	_, err = r.client.RbacV1().RoleBindings(namespace).Update(ctx, &roleBinding, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (r repo) GetRoleBinding(namespace string, name string) (rbacv1.RoleBinding, error) {
	ctx := context.TODO()
	resource, err := r.client.RbacV1().RoleBindings(namespace).Get(ctx, name, metav1.GetOptions{})
	return *resource, err
}
