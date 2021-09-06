package microservice

import (
	"context"
	"errors"
	"fmt"
	"log"

	v1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Gets the deployment that is linked to the microserviceID in the given namespace
func k8sGetDeployment(client *kubernetes.Clientset, context context.Context, namespace string, microserviceID string) (v1.Deployment, error) {
	deployments, err := client.AppsV1().Deployments(namespace).List(context, metaV1.ListOptions{})
	if err != nil {
		return v1.Deployment{}, err
	}

	found := false
	var foundDeployment v1.Deployment
	for _, deployment := range deployments.Items {
		_, ok := deployment.ObjectMeta.Labels["microservice"]
		if !ok {
			continue
		}

		if deployment.ObjectMeta.Annotations["dolittle.io/microservice-id"] == microserviceID {
			found = true
			foundDeployment = deployment
			break
		}
	}

	if !found {
		return v1.Deployment{}, errors.New("not-found")
	}
	return foundDeployment, nil
}

// Stops a deployment by scaling it down to zero
func k8sStopDeployment(client *kubernetes.Clientset, context context.Context, namespace string, deployment *v1.Deployment) error {
	s, err := client.AppsV1().Deployments(namespace).GetScale(context, deployment.Name, metaV1.GetOptions{})
	if err != nil {
		log.Fatal(err)
		return errors.New("issue")
	}

	sc := *s
	if sc.Spec.Replicas != 0 {
		sc.Spec.Replicas = 0
		_, err := client.AppsV1().Deployments(namespace).UpdateScale(context, deployment.Name, &sc, metaV1.UpdateOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("todo")
		}
	}
	return nil
}

// A utility method for handling errors when creating a kubernetes resource
func k8sHandleResourceCreationError(creationError error, onExists func()) error {
	if creationError != nil {
		if !k8serrors.IsAlreadyExists(creationError) {
			log.Fatal(creationError)
			return errors.New("issue")
		}
		onExists()
	}
	return nil
}

func k8sPrintAlreadyExists(resourceName string) {
	fmt.Printf("Skipping %s already exists\n", resourceName)
}
