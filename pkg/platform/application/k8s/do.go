package k8s

import (
	"context"
	"fmt"
	"log"

	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func Do(client kubernetes.Interface, resources Resources, k8sRepo platformK8s.K8sRepo) error {
	ctx := context.TODO()
	namespace := resources.Namespace.ObjectMeta.Name
	applicationID := resources.Namespace.Annotations["dolittle.io/application-id"]
	// Namespace

	_, err := client.CoreV1().Namespaces().Create(ctx, resources.Namespace, metav1.CreateOptions{})

	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
		}
		return nil
	}

	// Acr
	_, err = client.CoreV1().Secrets(namespace).Create(ctx, resources.Acr, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("err", err, resources.Acr.ObjectMeta.Name, namespace)
		deleteNamespace(client, namespace)
		return err
	}

	// Application Rbac
	_, err = client.RbacV1().Roles(namespace).Create(ctx, resources.Rbac.Role, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("err", err, resources.Rbac.Role.ObjectMeta.Name, namespace)
		deleteNamespace(client, namespace)
		return err
	}

	_, err = client.RbacV1().RoleBindings(namespace).Create(ctx, resources.Rbac.RoleBinding, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("err", err, resources.Rbac.RoleBinding.ObjectMeta.Name, namespace)
		deleteNamespace(client, namespace)
		return err
	}

	// Service accounts
	for _, serviceAccount := range resources.ServiceAccounts {
		err := k8sRepo.AddServiceAccount(serviceAccount.Name, serviceAccount.RoleBindingName, serviceAccount.Customer.ID, serviceAccount.Customer.Name, serviceAccount.Application.ID, serviceAccount.Application.Name)
		if err != nil {
			fmt.Println("err: service account setup", err, serviceAccount.Name, namespace)
			deleteNamespace(client, namespace)
			return err
		}
	}

	// Storage
	_, err = client.CoreV1().Secrets(namespace).Create(ctx, resources.Storage, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("err", err, resources.Storage.ObjectMeta.Name, namespace)
		deleteNamespace(client, namespace)
		return err
	}

	// Environments
	for _, environmentResource := range resources.Environments {
		// TODO where do we add the customer tenant?
		// tenants.json
		_, err = client.CoreV1().ConfigMaps(namespace).Create(context.TODO(), environmentResource.Tenants, metav1.CreateOptions{})
		if err != nil {
			fmt.Println("err", err, environmentResource.Tenants.ObjectMeta.Name, namespace)
			deleteNamespace(client, namespace)
			return err
		}

		// NetworkPolicy
		_, err = client.NetworkingV1().NetworkPolicies(namespace).Create(ctx, environmentResource.NetworkPolicy, metav1.CreateOptions{})
		if err != nil {
			fmt.Println("err", err, environmentResource.NetworkPolicy.ObjectMeta.Name, namespace)
			deleteNamespace(client, namespace)
			return err
		}

		// Mongo
		// Service
		_, err = client.CoreV1().Services(namespace).Create(context.TODO(), environmentResource.Mongo.Service, metav1.CreateOptions{})
		if err != nil {
			fmt.Println("err", err, environmentResource.Mongo.Service.ObjectMeta.Name, namespace)
			deleteNamespace(client, namespace)
			return err
		}

		// StatefulSet
		_, err = client.AppsV1().StatefulSets(namespace).Create(context.TODO(), environmentResource.Mongo.StatefulSet, metav1.CreateOptions{})
		if err != nil {
			fmt.Println("err", err, environmentResource.Mongo.Service.ObjectMeta.Name, namespace)
			deleteNamespace(client, namespace)
			return err
		}

		// Cronjob
		_, err = client.BatchV1beta1().CronJobs(namespace).Create(context.TODO(), environmentResource.Mongo.Cronjob, metav1.CreateOptions{})
		if err != nil {
			fmt.Println("err", err, environmentResource.Mongo.Cronjob.ObjectMeta.Name, namespace)
			deleteNamespace(client, namespace)
			return err
		}

		// Add specifc policy rules for this environment to the developer
		for _, policyRule := range environmentResource.RbacRolePolicyRules {
			err := k8sRepo.AddPolicyRule("developer", applicationID, policyRule)
			if err != nil {
				deleteNamespace(client, namespace)
				return err
			}
		}
	}

	// Create local-dev bindings for developers testing locally
	if resources.LocalDevRoleBindingToDeveloper != nil {
		_, err = client.RbacV1().RoleBindings(namespace).Create(context.TODO(), resources.LocalDevRoleBindingToDeveloper, metav1.CreateOptions{})
		if err != nil {
			fmt.Println("err", err, resources.LocalDevRoleBindingToDeveloper.ObjectMeta.Name, namespace)
			deleteNamespace(client, namespace)
			return err
		}
	}

	return nil
}

func deleteNamespace(client kubernetes.Interface, namespace string) {
	ctx := context.TODO()
	err := client.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
	if err != nil {
		log.Fatal(err)
	}
	// TODO maybe be less aggressive :P and call it undo add undo... slowly to an operator
}
