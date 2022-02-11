package k8s

import (
	"fmt"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
)

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
