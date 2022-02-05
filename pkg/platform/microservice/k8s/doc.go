package k8s

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/dolittle/platform-api/pkg/dolittle/k8s"
	v1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	TodoCustomersTenantID string = "17426336-fb8e-4425-8ab7-07d488367be9"
)

type MicroserviceK8sInfo struct {
	Tenant      k8s.Tenant
	Application k8s.Application
	Namespace   string
}

// K8sHasDeploymentWithName gets the microservice deployment that is has a specific name in the given namespace
func K8sHasDeploymentWithName(client kubernetes.Interface, context context.Context, namespace, name string) (bool, error) {
	deployments, err := client.AppsV1().Deployments(namespace).List(context, metav1.ListOptions{})
	if err != nil {
		return false, err
	}

	found := false
	for _, deployment := range deployments.Items {
		_, ok := deployment.ObjectMeta.Labels["microservice"]
		if !ok {
			continue
		}

		if deployment.Name == name {
			found = true
			break
		}
	}

	return found, nil
}

// Stops a deployment by scaling it down to zero
func K8sStopDeployment(client kubernetes.Interface, context context.Context, namespace string, deployment *v1.Deployment) error {
	s, err := client.AppsV1().Deployments(namespace).GetScale(context, deployment.Name, metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
		return errors.New("issue")
	}

	sc := *s
	if sc.Spec.Replicas != 0 {
		sc.Spec.Replicas = 0
		_, err := client.AppsV1().Deployments(namespace).UpdateScale(context, deployment.Name, &sc, metav1.UpdateOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("todo")
		}
	}
	return nil
}

// A utility method for handling errors when creating a kubernetes resource
func K8sHandleResourceCreationError(creationError error, onExists func()) error {
	if creationError != nil {
		if !k8serrors.IsAlreadyExists(creationError) {
			log.Fatal(creationError)
			return errors.New("issue")
		}
		onExists()
	}
	return nil
}

/// Finds and deletes all configmaps in namespace based on the given metav1.ListOptions
func K8sDeleteConfigmaps(client kubernetes.Interface, ctx context.Context, namespace string, listOpts metav1.ListOptions) error {
	configs, _ := client.CoreV1().ConfigMaps(namespace).List(ctx, listOpts)
	for _, config := range configs.Items {
		err := client.CoreV1().ConfigMaps(namespace).Delete(ctx, config.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("todo")
		}
	}
	return nil
}

/// Finds and deletes all secrets in namespace based on the given metav1.ListOptions
func K8sDeleteSecrets(client kubernetes.Interface, ctx context.Context, namespace string, listOpts metav1.ListOptions) error {
	secrets, _ := client.CoreV1().Secrets(namespace).List(ctx, listOpts)
	for _, secret := range secrets.Items {
		err := client.CoreV1().Secrets(namespace).Delete(ctx, secret.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("todo")
		}
	}
	return nil
}

/// Finds and deletes all ingresses in namespace based on the given metav1.ListOptions
func K8sDeleteIngresses(client kubernetes.Interface, ctx context.Context, namespace string, listOpts metav1.ListOptions) error {
	ingresses, _ := client.NetworkingV1().Ingresses(namespace).List(ctx, listOpts)
	for _, ingress := range ingresses.Items {
		err := client.NetworkingV1().Ingresses(namespace).Delete(ctx, ingress.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("issue")
		}
	}
	return nil
}

/// Finds and deletes all network policies in namespace based on the given metav1.ListOptions
func K8sDeleteNetworkPolicies(client kubernetes.Interface, ctx context.Context, namespace string, listOpts metav1.ListOptions) error {
	policies, _ := client.NetworkingV1().NetworkPolicies(namespace).List(ctx, listOpts)
	for _, policy := range policies.Items {
		err := client.NetworkingV1().NetworkPolicies(namespace).Delete(ctx, policy.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("issue")
		}
	}
	return nil
}

/// Finds and deletes all services in namespace based on the given metav1.ListOptions
func K8sDeleteServices(client kubernetes.Interface, ctx context.Context, namespace string, listOpts metav1.ListOptions) error {
	services, _ := client.CoreV1().Services(namespace).List(ctx, listOpts)
	for _, service := range services.Items {
		err := client.CoreV1().Services(namespace).Delete(ctx, service.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("issue")
		}
	}
	return nil
}

// Finds and deletes the deployment in the given namespace
func K8sDeleteDeployment(client kubernetes.Interface, ctx context.Context, namespace string, deployment *v1.Deployment) error {
	err := client.AppsV1().Deployments(namespace).Delete(ctx, deployment.Name, metav1.DeleteOptions{})
	if err != nil {
		log.Fatal(err)
		return errors.New("todo")
	}
	return nil
}

func K8sPrintAlreadyExists(resourceName string) {
	fmt.Printf("Skipping %s already exists\n", resourceName)
}

func DeleteRole(client kubernetes.Interface, ctx context.Context, namespace string, listOpts metav1.ListOptions) error {
	roles, _ := client.RbacV1().Roles(namespace).List(ctx, listOpts)
	for _, role := range roles.Items {
		err := client.RbacV1().Roles(namespace).Delete(ctx, role.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("issue")
		}
	}
	return nil
}

func DeleteRoleBindings(client kubernetes.Interface, ctx context.Context, namespace string, listOpts metav1.ListOptions) error {
	roleBindings, _ := client.RbacV1().RoleBindings(namespace).List(ctx, listOpts)
	for _, roleBinding := range roleBindings.Items {
		err := client.RbacV1().RoleBindings(namespace).Delete(ctx, roleBinding.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("issue")
		}
	}
	return nil
}
