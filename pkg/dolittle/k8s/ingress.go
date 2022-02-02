package k8s

import (
	"fmt"
	"strings"

	"github.com/docker/docker/pkg/namesgenerator"
	platform "github.com/dolittle/platform-api/pkg/platform"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewDevelopmentCustomerTenantInfo(environment string, indexID int, microserviceID string) platform.CustomerTenantInfo {
	return NewCustomerTenantInfo(environment, indexID, microserviceID, platform.DevelopmentCustomersTenantID)
}

// TODO remove indexID, we are not using it
func NewCustomerTenantInfo(environment string, indexID int, microserviceID string, customerTenantID string) platform.CustomerTenantInfo {
	// TODO how do we make sure the host is not already in use
	// Do we look it up in the cluster (source of truth)
	// Do we rely on the data in git?
	// My gut says cluster
	return platform.CustomerTenantInfo{
		Environment:      environment,
		CustomerTenantID: customerTenantID,
		Ingress:          NewCustomerTenantIngress(),
		MicroservicesRel: []platform.CustomerTenantMicroserviceRel{
			{
				MicroserviceID: microserviceID,
				Hash:           ResourcePrefix(microserviceID, customerTenantID),
			},
		},
		RuntimeInfo: platform.CustomerTenantRuntimeStorageInfo{
			DatabasePrefix: ResourcePrefix(microserviceID, customerTenantID),
		},
	}
}

func NewCustomerTenantIngress() platform.CustomerTenantIngress {
	domainPrefix := namesgenerator.GetRandomName(-1)
	domainPrefix = strings.ReplaceAll(domainPrefix, "_", "-")

	host := fmt.Sprintf("%s.dolittle.cloud", domainPrefix)
	secretName := fmt.Sprintf("%s-certificate", domainPrefix)

	return platform.CustomerTenantIngress{
		Host:         host,
		DomainPrefix: domainPrefix,
		SecretName:   secretName,
	}
}

func NewMicroserviceIngressWithEmptyRules(platformEnvironment string, microservice Microservice) *networkingv1.Ingress {
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
	if platformEnvironment == "prod" {
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
