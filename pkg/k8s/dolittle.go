package k8s

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type repo struct {
	client     kubernetes.Interface
	logContext logrus.FieldLogger
}

type Repo interface {
	GetIngressesByEnvironmentWithMicoservices(namespace string, environment string) ([]networkingv1.Ingress, error)
}

func NewRepo(client kubernetes.Interface, logContext logrus.FieldLogger) Repo {
	return repo{
		client:     client,
		logContext: logContext,
	}
}

func (r repo) GetIngressesByEnvironmentWithMicoservices(namespace string, environment string) ([]networkingv1.Ingress, error) {
	ctx := context.TODO()
	opts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("tenant,environment=%s,microservice", environment),
	}

	items, err := r.client.NetworkingV1().Ingresses(namespace).List(ctx, opts)
	return items.Items, err
}
