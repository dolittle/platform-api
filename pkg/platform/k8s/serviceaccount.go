package k8s

import (
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewServiceAccountResource(serviceAccountName string, customerID string, customerName string, applicationID string, applicationName string) corev1.ServiceAccount {
	namespace := GetApplicationNamespace(applicationID)
	labels := GetLabelsForApplication(customerName, applicationName)
	annotations := GetAnnotationsForApplication(customerID, applicationID)
	return corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        serviceAccountName,
			Namespace:   namespace,
			Annotations: annotations,
			Labels:      labels,
		},
	}
}

func (r *K8sRepo) AddServiceAccount(serviceAccount string, roleBinding string, customerID string, customerName string, applicationID string, applicationName string) error {
	// TODO can we make the serviceaccount == roleBinding and not include the 2nd parameter?
	logContext := r.logContext.WithFields(logrus.Fields{
		"customerID":     customerID,
		"applicationID":  applicationID,
		"serviceAccount": serviceAccount,
		"roleBinding":    roleBinding,
		"function":       "createServiceAccount",
	})

	resource := NewServiceAccountResource(serviceAccount, customerID, customerName, applicationID, applicationName)
	_, err := r.CreateServiceAccountFromResource(logContext, &resource)

	if err != nil && err != ErrAlreadyExists {
		return err
	}

	roleBindingResource := NewRoleBindingWithoutSubjects(roleBinding, "developer", customerID, customerName, applicationID, applicationName)
	_, err = r.CreateRoleBindingFromResource(logContext, &roleBindingResource)

	if err != nil && err != ErrAlreadyExists {
		return err
	}

	_, err = r.AddServiceAccountToRoleBinding(logContext, applicationID, roleBinding, serviceAccount)
	if err != nil {
		return err
	}
	return nil
}
