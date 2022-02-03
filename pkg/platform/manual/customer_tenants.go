package manual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/automate"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r Repo) GetIngressesByEnvironment(namespace string, environment string) (*networkingv1.IngressList, error) {
	ctx := context.TODO()
	opts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("tenant,environment=%s,microservice", environment),
	}

	return r.client.NetworkingV1().Ingresses(namespace).List(ctx, opts)
}

func (r Repo) GetIngressessByCustomerTenantID(ingresses *networkingv1.IngressList, customerTenantID string) ([]networkingv1.Ingress, error) {
	filtered := funk.Filter(ingresses.Items, func(ingress networkingv1.Ingress) bool {
		tenantHeaderAnnotation := ingress.GetObjectMeta().GetAnnotations()["nginx.ingress.kubernetes.io/configuration-snippet"]
		proxyHeaderTenantID := platformK8s.GetCustomerTenantIDFromNginxConfigurationSnippet(tenantHeaderAnnotation)
		return proxyHeaderTenantID == customerTenantID
	}).([]networkingv1.Ingress)

	uniq := make([]networkingv1.Ingress, 0)
	for _, ingress := range filtered {
		match := funk.Contains(uniq, func(current networkingv1.Ingress) bool {
			hostA := current.Spec.TLS[0].Hosts[0]
			hostB := ingress.Spec.TLS[0].Hosts[0]
			return hostA == hostB
		})

		if match {
			continue
		}
		uniq = append(uniq, ingress)
	}

	if len(uniq) == 0 {
		return []networkingv1.Ingress{}, errors.New("not-found")
	}

	return uniq, nil
}

func (r Repo) GetCustomerTenantIngresses(ingresses *networkingv1.IngressList, customerTenantID string, logContext logrus.FieldLogger) []platform.CustomerTenantIngress {
	filtered, err := r.GetIngressessByCustomerTenantID(ingresses, customerTenantID)

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to find one ingress")
		return []platform.CustomerTenantIngress{}
	}

	items := make([]platform.CustomerTenantIngress, 0)
	for _, ingress := range filtered {
		microserviceID := ingress.Annotations["dolittle.io/microservice-id"]

		host := ingress.Spec.TLS[0].Hosts[0]
		secretName := ingress.Spec.TLS[0].SecretName
		domainPrefix := "na"

		for _, rule := range ingress.Spec.Rules {
			for _, ingressPath := range rule.HTTP.Paths {
				item := platform.CustomerTenantIngress{
					MicroserviceID: microserviceID,
					Host:           host,
					SecretName:     secretName,
					DomainPrefix:   domainPrefix,
					Path:           ingressPath.Path,
				}
				items = append(items, item)
			}
		}
	}

	return items
}

func (r Repo) GetCustomerTenantIDSByEnvironment(namespace string, environment string) []string {
	ctx := context.TODO()
	client := r.client

	configMaps, err := automate.GetCustomerTenantsConfigMaps(ctx, client, namespace)
	if err != nil {
		panic(err)
	}

	var runtimeTenants platform.RuntimeTenantsIDS
	for _, configMap := range configMaps {
		if configMap.Labels["environment"] != environment {
			continue
		}

		json.Unmarshal([]byte(configMap.Data["tenants.json"]), &runtimeTenants)

		break

	}

	keys := make([]string, 0, len(runtimeTenants))
	for k := range runtimeTenants {
		keys = append(keys, k)
	}
	return keys
}

func (r Repo) GetCustomerTenants(ctx context.Context, namespace string) []platform.CustomerTenantInfo {
	client := r.client
	items := make([]platform.CustomerTenantInfo, 0)

	//Get customerTenants
	// Get Environments
	customerTenantsConfigMaps, err := automate.GetCustomerTenantsConfigMaps(ctx, client, namespace)
	if err != nil {
		panic(err)
	}

	dolittleConfigMaps, err := automate.GetDolittleConfigMaps(ctx, client, namespace)
	if err != nil {
		r.logContext.WithFields(logrus.Fields{
			"namespace": namespace,
		}).Fatal("Failed to get *-dolittle configmaps")
	}

	environments := make([]string, 0)
	for _, configMap := range customerTenantsConfigMaps {
		environments = append(environments, configMap.Labels["environment"])
	}

	for _, environment := range environments {

		customerTenantIDS := r.GetCustomerTenantIDSByEnvironment(namespace, environment)

		ingresses, err := r.GetIngressesByEnvironment(namespace, environment)
		if err != nil {
			panic(err)
		}

		filteredDolittleConfigMaps := funk.Filter(dolittleConfigMaps, func(configMap corev1.ConfigMap) bool {
			return configMap.Labels["environment"] == environment
		}).([]corev1.ConfigMap)

		for _, customerTenantID := range customerTenantIDS {
			logContext := r.logContext.WithFields(logrus.Fields{
				"customer_tenant_id": customerTenantID,
				"environment":        environment,
				"namespace":          namespace,
			})

			item := platform.CustomerTenantInfo{
				Environment:      environment,
				CustomerTenantID: customerTenantID,
				Ingresses:        []platform.CustomerTenantIngress{},
				MicroservicesRel: []platform.CustomerTenantMicroserviceRel{},
				//RuntimeInfo:      platform.CustomerTenantRuntimeStorageInfo{},
			}

			item.Ingresses = r.GetCustomerTenantIngresses(ingresses, customerTenantID, logContext)
			item.MicroservicesRel = r.GetCustomerTenantMicroserviceRelationships(filteredDolittleConfigMaps, customerTenantID, logContext)
			items = append(items, item)
		}

	}
	return items
}

// GetCustomerTenantMicroserviceRelationships
// This only finds within one namespace, to make it find across namespaces we would need to know which namespaces to check
func (r Repo) GetCustomerTenantMicroserviceRelationships(configMaps []corev1.ConfigMap, customerTenantID string, logContext logrus.FieldLogger) []platform.CustomerTenantMicroserviceRel {
	// TODO would be nice to query the runtime for this

	relationships := make([]platform.CustomerTenantMicroserviceRel, 0)
	for _, configMap := range configMaps {
		logContext := r.logContext.WithFields(logrus.Fields{
			"configmap_name": configMap.Name,
		})

		var microserviceResources dolittleK8s.MicroserviceResources
		err := json.Unmarshal([]byte(configMap.Data["resources.json"]), &microserviceResources)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"error": err,
			}).Error("Failed to parse resources.json")
			continue
		}
		_, ok := microserviceResources[customerTenantID]
		if !ok {
			continue
		}
		microserviceID := configMap.Annotations["dolittle.io/microservice-id"]

		relationships = append(relationships, platform.CustomerTenantMicroserviceRel{
			MicroserviceID: microserviceID,
			Hash:           dolittleK8s.ResourcePrefix(microserviceID, customerTenantID),
		})
	}
	return relationships
}

// GetCustomerTenantMicroserviceResource
// Given a dolittleConfigMap return the customerTenant block
// Not in use, but could be.
func (r Repo) GetCustomerTenantMicroserviceResource(dolittleConfigMap corev1.ConfigMap, customerTenantID string) (dolittleK8s.MicroserviceResource, error) {
	logContext := r.logContext.WithFields(logrus.Fields{
		"customer_tenant_id": customerTenantID,
		"microservice_id":    dolittleConfigMap.Annotations["dolittle.io/microservice-id"],
		"configmap_name":     dolittleConfigMap.Name,
		"environment":        dolittleConfigMap.Labels["environment"],
		"namespace":          dolittleConfigMap.Namespace,
	})

	var microserviceResources dolittleK8s.MicroserviceResources
	err := json.Unmarshal([]byte(dolittleConfigMap.Data["resources.json"]), &microserviceResources)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to parse resources.json")

		return dolittleK8s.MicroserviceResource{}, errors.New("bad-json")
	}

	// Find the key
	item, ok := microserviceResources[customerTenantID]
	if !ok {
		err := errors.New("customer-tenant-not-found")
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to find customer tenant")

		return dolittleK8s.MicroserviceResource{}, err
	}

	return item, nil
}
