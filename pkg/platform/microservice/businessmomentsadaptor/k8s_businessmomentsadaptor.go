package businessmomentsadaptor

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type K8sRepo struct {
	k8sClient *kubernetes.Clientset
}

func NewK8sRepo(k8sClient *kubernetes.Clientset) Repo {
	return K8sRepo{
		k8sClient: k8sClient,
	}
}

func NewBusinessMomentsConfigmap(microservice k8s.Microservice) *coreV1.ConfigMap {
	name := fmt.Sprintf("%s-%s-business-moments",
		microservice.Environment,
		microservice.Name,
	)

	labels := k8s.GetLabels(microservice)
	annotations := k8s.GetAnnotations(microservice)

	name = strings.ToLower(name)

	return &coreV1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: annotations,
			Labels:      labels,
			Namespace:   fmt.Sprintf("application-%s", microservice.Application.ID),
		},
	}
}

func (r K8sRepo) GetBusinessMomentsConfigmap(
	applicationID string,
	environment string,
	microserviceID string,
) (coreV1.ConfigMap, error) {
	ctx := context.TODO()
	client := r.k8sClient
	namespace := fmt.Sprintf("application-%s", applicationID)
	items, err := client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})

	if err != nil {
		return coreV1.ConfigMap{}, err
	}

	found := false
	var foundConfigmap v1.ConfigMap

	for _, item := range items.Items {

		annotationsMap := item.GetObjectMeta().GetAnnotations()
		labelMap := item.GetObjectMeta().GetLabels()

		if annotationsMap["dolittle.io/application-id"] != applicationID {
			continue
		}

		if strings.ToLower(labelMap["environment"]) != environment {
			continue
		}

		if annotationsMap["dolittle.io/microservice-id"] != microserviceID {
			continue
		}

		found = true
		foundConfigmap = item
		break
	}

	if !found {
		// TO make we need
		return coreV1.ConfigMap{}, platform.ErrNotFound
	}
	return foundConfigmap, nil
}

func (r K8sRepo) SaveBusinessMomentsConfigmap(newConfigmap coreV1.ConfigMap, data []byte) error {
	ctx := context.TODO()
	client := r.k8sClient
	annotationsMap := newConfigmap.GetObjectMeta().GetAnnotations()
	applicationID, ok := annotationsMap["dolittle.io/application-id"]
	if !ok {
		return errors.New("Missing application ID")
	}

	namespace := fmt.Sprintf("application-%s", applicationID)
	newConfigmap.Data = map[string]string{
		"businessmoments.json": string(data),
	}

	_, err := client.CoreV1().ConfigMaps(namespace).Update(ctx, &newConfigmap, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}
