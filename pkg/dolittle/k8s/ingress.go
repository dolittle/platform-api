package k8s

import (
	"fmt"
	"strings"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewIngress(microservice Microservice) *networkingv1.Ingress {
	ingressName := fmt.Sprintf("%s-%s",
		microservice.Environment,
		microservice.Name,
	)

	className := "nginx"

	labels := GetLabels(microservice)
	annotations := GetAnnotations(microservice)
	annotations["cert-manager.io/cluster-issuer"] = "letsencrypt-production"
	annotations["nginx.ingress.kubernetes.io/configuration-snippet"] = fmt.Sprintf("proxy_set_header Tenant-ID \"%s\";\n", microservice.ResourceID)

	ingressName = strings.ToLower(ingressName)

	return &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.k8s.io/v1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        ingressName,
			Annotations: annotations,
			Labels:      labels,
			Namespace:   fmt.Sprintf("application-%s", microservice.Application.ID),
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &className,
			Rules:            []networkingv1.IngressRule{},
		},
	}
}

func AddIngressTLS(hosts []string, secretName string) []networkingv1.IngressTLS {
	return []networkingv1.IngressTLS{
		{
			Hosts:      hosts,
			SecretName: secretName,
		},
	}
}

func AddIngressRule(host string, paths []SimpleIngressRule) networkingv1.IngressRule {

	ingressPaths := []networkingv1.HTTPIngressPath{}
	for _, ingressPath := range paths {
		ingressPaths = append(ingressPaths, networkingv1.HTTPIngressPath{
			Path:     ingressPath.Path,
			PathType: &ingressPath.PathType,
			Backend: networkingv1.IngressBackend{
				Service: &networkingv1.IngressServiceBackend{
					Name: ingressPath.ServiceName,
					Port: networkingv1.ServiceBackendPort{
						Name: ingressPath.ServicePortName,
					},
				},
			},
		})
	}
	return networkingv1.IngressRule{
		Host: host,
		IngressRuleValue: networkingv1.IngressRuleValue{
			HTTP: &networkingv1.HTTPIngressRuleValue{
				Paths: ingressPaths,
			},
		},
	}
}
