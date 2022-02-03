package k8s

import (
	"fmt"
	"strings"

	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO This is not in use GetMicroserviceRbacSubjects, remove
func GetMicroserviceRbacSubjects(azureGroupID string, tenantGroup string, extraSubjects ...rbacv1.Subject) []rbacv1.Subject {
	subjects := []rbacv1.Subject{
		{
			Kind:     "Group",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     azureGroupID,
		},
		{
			Kind:     "Group",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     tenantGroup,
		},
	}

	subjects = append(subjects, extraSubjects...)
	return subjects
}

// TODO this is not in use anymmore, remove
// NewMicroserviceRbac
func NewMicroserviceRbac(microserviceName string, microserviceID string, microserviceKind string, tenant dolittleK8s.Tenant, application dolittleK8s.Application, environment string, subjects []rbacv1.Subject) (role *rbacv1.Role, roleBinding *rbacv1.RoleBinding) {
	// TODO the microserviceName might need to be internal and shorter
	namespace := fmt.Sprintf("application-%s", application.ID)
	labels := platformK8s.GetLabelsForMicroservice(tenant.Name, application.Name, environment, microserviceName)
	annotations := platformK8s.GetAnnotationsForMicroservice(tenant.ID, application.ID, microserviceID, microserviceKind)

	prefix := strings.ToLower(fmt.Sprintf("%s-%s", environment, microserviceName))
	name := strings.ToLower(fmt.Sprintf("developer-%s", prefix))
	rules := NewMicroservicePolicyRules(microserviceName, environment)
	role = &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/rbacv1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Rules: rules,
	}

	roleBinding = &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/rbacv1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Subjects: subjects,
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     name,
		},
	}
	return role, roleBinding
}

func NewMicroservicePolicyRules(microserviceName string, environment string) []rbacv1.PolicyRule {
	prefix := strings.ToLower(fmt.Sprintf("%s-%s", environment, microserviceName))

	configMaps := []string{
		fmt.Sprintf("%s-env-variables", prefix),
		fmt.Sprintf("%s-config-files", prefix),
	}

	// TODO do we really need secret when its in a secret?
	secrets := []string{
		fmt.Sprintf("%s-secret-env-variables", prefix),
	}

	return []rbacv1.PolicyRule{
		{
			Verbs: []string{
				"get",
				"patch",
			},
			APIGroups:     []string{""},
			Resources:     []string{"configmaps"},
			ResourceNames: configMaps,
		},
		{
			Verbs: []string{
				"get",
				"patch",
			},
			APIGroups:     []string{""},
			Resources:     []string{"secrets"},
			ResourceNames: secrets,
		},
	}

}
