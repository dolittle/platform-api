package k8s

import (
	"fmt"
	"strings"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewNetworkPolicy(microservice Microservice) *networkingv1.NetworkPolicy {
	name := fmt.Sprintf("%s-%s-ingress",
		microservice.Environment,
		microservice.Name,
	)
	name = strings.ToLower(name)

	labels := GetLabels(microservice)
	annotations := GetAnnotations(microservice)

	return &networkingv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.k8s.io/v1",
			Kind:       "NetworkPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: annotations,
			Labels:      labels,
			Namespace:   fmt.Sprintf("application-%s", microservice.Application.ID),
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: labels,
			},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"system": "Ingress",
								},
							},
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"system":  "Ingress",
									"ingress": "Nginx",
								},
							},
						},
					},
				},
			},
		},
	}
}
