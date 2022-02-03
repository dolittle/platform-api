package customertenant

import (
	"fmt"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	networkingv1 "k8s.io/api/networking/v1"
)

func CreateIngresses(platformEnvironment string, customerTenants []platform.CustomerTenantInfo, microservice dolittleK8s.Microservice, rules []dolittleK8s.SimpleIngressRule) []*networkingv1.Ingress {
	// TODO why *?
	ingresses := make([]*networkingv1.Ingress, 0)
	for _, customerTenant := range customerTenants {
		for _, ingressConfig := range customerTenant.Ingresses {
			ingress := dolittleK8s.NewMicroserviceIngressWithEmptyRules(platformEnvironment, microservice)
			// TODO should customerTenant.CustomerTenantID[0:7] be a hash?
			newName := fmt.Sprintf("%s-%s", ingress.ObjectMeta.Name, customerTenant.CustomerTenantID[0:7])
			ingress.ObjectMeta.Name = newName
			ingress = dolittleK8s.AddCustomerTenantIDToIngress(customerTenant.CustomerTenantID, ingress)
			ingress.Spec.TLS = dolittleK8s.AddIngressTLS([]string{ingressConfig.Host}, ingressConfig.SecretName)
			ingress.Spec.Rules = append(ingress.Spec.Rules, dolittleK8s.AddIngressRule(ingressConfig.Host, rules))
			ingresses = append(ingresses, ingress)
		}
	}
	return ingresses
}
