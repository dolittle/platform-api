package customertenant

import (
	"fmt"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	networkingv1 "k8s.io/api/networking/v1"
)

func CreateIngresses(isProduction bool, customerTenants []platform.CustomerTenantInfo, microservice dolittleK8s.Microservice, serviceName string, ingressInfo platform.HttpInputSimpleIngress) []*networkingv1.Ingress {
	rules := []dolittleK8s.SimpleIngressRule{
		{
			Path:            ingressInfo.Path,
			PathType:        networkingv1.PathType(ingressInfo.Pathtype),
			ServiceName:     serviceName,
			ServicePortName: "http",
		},
	}

	ingresses := make([]*networkingv1.Ingress, 0)
	for _, customerTenant := range customerTenants {
		for _, config := range customerTenant.Hosts {
			// At this point we are assumed secret name is correct
			ingress := dolittleK8s.NewMicroserviceIngressWithEmptyRules(isProduction, microservice)
			newName := fmt.Sprintf("%s-%s", ingress.ObjectMeta.Name, customerTenant.CustomerTenantID[0:7])
			ingress.ObjectMeta.Name = newName
			ingress = dolittleK8s.AddCustomerTenantIDToIngress(customerTenant.CustomerTenantID, ingress)
			ingress.Spec.TLS = dolittleK8s.AddIngressTLS([]string{config.Host}, config.SecretName)
			ingress.Spec.Rules = append(ingress.Spec.Rules, dolittleK8s.AddIngressRule(config.Host, rules))
			ingresses = append(ingresses, ingress)
		}
	}
	return ingresses
}
