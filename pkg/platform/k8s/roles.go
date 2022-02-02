package k8s

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewRoleBindingWithoutSubjects(roleBindingName string, roleName string, customerID string, customerName string, applicationID string, applicationName string) rbacv1.RoleBinding {
	namespace := GetApplicationNamespace(applicationID)
	labels := GetLabelsForApplication(customerName, applicationName)
	annotations := GetAnnotationsForApplication(customerID, applicationID)

	return rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/rbacv1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        roleBindingName,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     roleName,
		},
	}
}
