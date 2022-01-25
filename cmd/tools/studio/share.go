package studio

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/dolittle/platform-api/pkg/platform"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func extractApplications(ctx context.Context, client kubernetes.Interface) []platform.HttpResponseApplication {
	applications := make([]platform.HttpResponseApplication, 0)

	for _, namespace := range getNamespaces(ctx, client) {
		if isApplicationNamespace(namespace) {
			applications = append(applications, getApplicationFromK8s(ctx, client, namespace))
		}
	}

	return applications
}

func getNamespaces(ctx context.Context, client kubernetes.Interface) []corev1.Namespace {
	namespacesList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	return namespacesList.Items
}

func isApplicationNamespace(namespace corev1.Namespace) bool {
	if !strings.HasPrefix(namespace.GetName(), "application-") {
		return false
	}
	if _, hasAnnotation := namespace.Annotations["dolittle.io/tenant-id"]; !hasAnnotation {
		return false
	}
	if _, hasAnnotation := namespace.Annotations["dolittle.io/application-id"]; !hasAnnotation {
		return false
	}
	if _, hasLabel := namespace.Labels["tenant"]; !hasLabel {
		return false
	}
	if _, hasLabel := namespace.Labels["application"]; !hasLabel {
		return false
	}

	return true
}

func getApplicationFromK8s(ctx context.Context, client kubernetes.Interface, namespace corev1.Namespace) platform.HttpResponseApplication {
	application := platform.HttpResponseApplication{
		ID:         namespace.Annotations["dolittle.io/application-id"],
		Name:       namespace.Labels["application"],
		TenantID:   namespace.Annotations["dolittle.io/tenant-id"],
		TenantName: namespace.Labels["tenant"],
	}

	application.Environments = getApplicationEnvironmentsFromK8s(ctx, client, namespace.GetName(), application.ID, application.TenantID)

	return application
}

func getApplicationEnvironmentsFromK8s(ctx context.Context, client kubernetes.Interface, namespace, applicationID, tenantID string) []platform.HttpInputEnvironment {
	environments := make([]platform.HttpInputEnvironment, 0)
	for _, configmap := range getConfigmaps(ctx, client, namespace) {
		if isEnvironmentTenantsConfig(configmap) {
			environment := platform.HttpInputEnvironment{
				Name:          configmap.Labels["environment"],
				TenantID:      tenantID,
				ApplicationID: applicationID,
			}

			environmentLabels := configmap.Labels

			environment.Tenants = getTenantsFromTenantsJson(configmap.Data["tenants.json"])
			environment.Ingresses = getEnvironmentIngressesFromK8s(ctx, client, namespace, environmentLabels)

			environments = append(environments, environment)
		}
	}
	return environments
}

func getConfigmaps(ctx context.Context, client kubernetes.Interface, namespace string) []corev1.ConfigMap {
	configmapList, err := client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	return configmapList.Items
}

func isEnvironmentTenantsConfig(configmap corev1.ConfigMap) bool {
	if _, hasAnnotation := configmap.Annotations["dolittle.io/tenant-id"]; !hasAnnotation {
		return false
	}
	if _, hasAnnotation := configmap.Annotations["dolittle.io/application-id"]; !hasAnnotation {
		return false
	}
	if _, hasLabel := configmap.Labels["tenant"]; !hasLabel {
		return false
	}
	if _, hasLabel := configmap.Labels["application"]; !hasLabel {
		return false
	}
	if _, hasLabel := configmap.Labels["environment"]; !hasLabel {
		return false
	}

	_, hasTenantsJson := configmap.Data["tenants.json"]
	return hasTenantsJson
}

func getTenantsFromTenantsJson(tenantsJsonContent string) []platform.TenantId {
	tenantObjects := make(map[string]interface{})
	if err := json.Unmarshal([]byte(tenantsJsonContent), &tenantObjects); err != nil {
		panic(err.Error())
	}

	tenants := make([]platform.TenantId, 0)
	for tenantID := range tenantObjects {
		tenants = append(tenants, platform.TenantId(tenantID))
	}
	return tenants
}

func getEnvironmentIngressesFromK8s(ctx context.Context, client kubernetes.Interface, namespace string, environmentLabels labels.Set) platform.EnvironmentIngresses {
	ingresses := make(map[platform.TenantId]platform.EnvironmentIngress)
	for _, ingress := range getIngresses(ctx, client, namespace, environmentLabels) {
		if !isMicroserviceIngress(ingress) {
			continue
		}

		tenantIDFound, tenantID := tryGetTenantFromIngress(ingress)
		if !tenantIDFound {
			continue
		}

		for _, rule := range ingress.Spec.Rules {
			host := rule.Host
			domainPrefix := strings.TrimSuffix(host, ".dolittle.cloud")

			secretNameFound, secretName := tryGetIngressSecretNameForHost(ingress, host)

			if secretNameFound {
				environmentIngress := platform.EnvironmentIngress{
					Host:         host,
					DomainPrefix: domainPrefix,
					SecretName:   secretName,
				}

				ingresses[tenantID] = environmentIngress
				break
			}
		}
	}
	return ingresses
}

func getIngresses(ctx context.Context, client kubernetes.Interface, namespace string, environmentLabels labels.Set) []netv1.Ingress {
	ingressList, err := client.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.FormatLabels(environmentLabels),
	})
	if err != nil {
		panic(err.Error())
	}
	return ingressList.Items
}

func isMicroserviceIngress(ingress netv1.Ingress) bool {
	if _, hasAnnotation := ingress.Annotations["dolittle.io/tenant-id"]; !hasAnnotation {
		return false
	}
	if _, hasAnnotation := ingress.Annotations["dolittle.io/application-id"]; !hasAnnotation {
		return false
	}
	if _, hasAnnotation := ingress.Annotations["dolittle.io/microservice-id"]; !hasAnnotation {
		return false
	}
	if _, hasLabel := ingress.Labels["tenant"]; !hasLabel {
		return false
	}
	if _, hasLabel := ingress.Labels["application"]; !hasLabel {
		return false
	}
	if _, hasLabel := ingress.Labels["environment"]; !hasLabel {
		return false
	}
	if _, hasLabel := ingress.Labels["microservice"]; !hasLabel {
		return false
	}

	return true
}

var tenantFromIngressAnnotationExtractor = regexp.MustCompile(`proxy_set_header\s+Tenant-ID\s+"([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})"`)

func tryGetTenantFromIngress(ingress netv1.Ingress) (bool, platform.TenantId) {
	tenantHeaderAnnotation := ingress.GetObjectMeta().GetAnnotations()["nginx.ingress.kubernetes.io/configuration-snippet"]
	tenantID := tenantFromIngressAnnotationExtractor.FindStringSubmatch(tenantHeaderAnnotation)
	if tenantID == nil {
		return false, ""
	}
	return true, platform.TenantId(tenantID[1])
}

func tryGetIngressSecretNameForHost(ingress netv1.Ingress, host string) (bool, string) {
	for _, tlsConfig := range ingress.Spec.TLS {
		for _, tlsHost := range tlsConfig.Hosts {
			if tlsHost == host {
				return true, tlsConfig.SecretName
			}
		}
	}
	return false, ""
}
