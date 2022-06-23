package k8s

import (
	"fmt"
	"strings"

	"github.com/docker/docker/pkg/namesgenerator"
	platform "github.com/dolittle/platform-api/pkg/platform"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewDevelopmentCustomerTenantInfo(environment string, microserviceID string) platform.CustomerTenantInfo {
	return NewCustomerTenantInfo(environment, microserviceID, platform.DevelopmentCustomersTenantID)
}

func NewCustomerTenantInfo(environment string, microserviceID string, customerTenantID string) platform.CustomerTenantInfo {
	// TODO https://app.asana.com/0/1201325052247030/1201777908041868/f
	return platform.CustomerTenantInfo{
		Environment:      environment,
		CustomerTenantID: customerTenantID,
		Hosts: []platform.CustomerTenantHost{
			NewCustomerTenantHost(microserviceID),
		},
		MicroservicesRel: []platform.CustomerTenantMicroserviceRel{
			{
				MicroserviceID: microserviceID,
				Hash:           ResourcePrefix(microserviceID, customerTenantID),
			},
		},
	}
}

func NewCustomerTenantHost(microserviceID string) platform.CustomerTenantHost {
	domainPrefix := namesgenerator.GetRandomName(0)
	domainPrefix = strings.ReplaceAll(domainPrefix, "_", "-")

	host := fmt.Sprintf("%s.dolittle.cloud", domainPrefix)
	secretName := fmt.Sprintf("%s-certificate", domainPrefix)
	return platform.CustomerTenantHost{
		Host:       host,
		SecretName: secretName,
	}
}

// NewMicroserviceIngressWithEmptyRules
// We use cert-manager and by default it is set to staging
func NewMicroserviceIngressWithEmptyRules(isProduction bool, microservice Microservice) *networkingv1.Ingress {
	namespace := fmt.Sprintf("application-%s", microservice.Application.ID)
	ingressName := fmt.Sprintf("%s-%s",
		microservice.Environment,
		microservice.Name,
	)
	ingressName = strings.ToLower(ingressName)

	className := "nginx"
	labels := GetLabels(microservice)
	annotations := GetAnnotations(microservice)

	// TODO might not work locally, you wont get https
	annotations["cert-manager.io/cluster-issuer"] = "letsencrypt-staging"
	if isProduction {
		annotations["cert-manager.io/cluster-issuer"] = "letsencrypt-production"

	}

	return &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.k8s.io/v1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        ingressName,
			Annotations: annotations,
			Labels:      labels,
			Namespace:   namespace,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &className,
			Rules:            []networkingv1.IngressRule{},
		},
	}
}

func AddCustomerTenantIDToIngress(customerTenantID string, ingress *networkingv1.Ingress) *networkingv1.Ingress {
	// TODO I wonder why we need the end
	ingress.Annotations["nginx.ingress.kubernetes.io/configuration-snippet"] = fmt.Sprintf("proxy_set_header Tenant-ID \"%s\";\n", customerTenantID)
	return ingress
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
