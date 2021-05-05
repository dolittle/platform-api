package microservice

// TODO move this  to shared
import (
	"context"
	"errors"
	"fmt"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Tenant struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type Application struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Tenant Tenant `json:"tenant"`
}

type k8sRepo struct {
	k8sClient *kubernetes.Clientset
}

func NewK8sRepo(k8sClient *kubernetes.Clientset) k8sRepo {
	return k8sRepo{
		k8sClient: k8sClient,
	}
}

//annotations:
//    dolittle.io/tenant-id: 388c0cc7-24b2-46a7-8735-b583ce21e01b
//    dolittle.io/application-id: c52e450e-4877-47bf-a584-7874c205e2b9
//  labels:
//    tenant: Flokk
//    application: Shepherd

func (r *k8sRepo) GetIngress(applicationID string) (string, error) {
	ctx := context.TODO()
	opts := metaV1.ListOptions{
		LabelSelector: "",
	}

	namespace := fmt.Sprintf("application-%s", applicationID)
	ingresses, _ := r.k8sClient.NetworkingV1().Ingresses(namespace).List(ctx, opts)
	for _, ingress := range ingresses.Items {
		if len(ingress.Spec.Rules) > 0 {
			return ingress.Spec.Rules[0].Host, nil
		}
	}

	return "", errors.New("")
}

func (r *k8sRepo) GetApplication(applicationID string) (Application, error) {
	client := r.k8sClient
	ctx := context.TODO()
	name := fmt.Sprintf("application-%s", applicationID)
	ns, err := client.CoreV1().Namespaces().Get(ctx, name, metaV1.GetOptions{})

	if err != nil {
		return Application{}, err
	}

	annotationsMap := ns.GetObjectMeta().GetAnnotations()
	labelMap := ns.GetObjectMeta().GetLabels()

	return Application{
		Name: labelMap["application"],
		ID:   annotationsMap["dolittle.io/application-id"],
		Tenant: Tenant{
			Name: labelMap["tenant"],
			ID:   annotationsMap["dolittle.io/tenant-id"],
		},
	}, nil
}
