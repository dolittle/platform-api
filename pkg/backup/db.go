package backup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/utils"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type AzureStorageInfo struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type DolittleTenant struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type DolittleApplication struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Tenant         DolittleTenant         `json:"tenant"`
	Environment    string                 `json:"environment"`
	Microservices  []DolittleMicroservice `json:"microservices"`
	TenantIDS      []string               `json:"tenant_ids"`
	AzureShareName string                 `json:"azure_share_name"`
	IngressHosts   []string               `json:"ingress_hosts"`
}

type DolittleMicroservice struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Application string `json:"application"`
	Environment string `json:"environment"`
}

type DolittleCustomerTenant struct {
	Namespace           string           `json:"namespace"`
	ApplicationID       string           `json:"application_id"`
	Tenant              DolittleTenant   `json:"tenant"`
	AzureStorageAccount AzureStorageInfo `json:"azure_storage_account"`

	Applications []DolittleApplication `json:"applications"`
}

func Rebuild(kubeconfig string, withSecrets bool) []DolittleCustomerTenant {
	ctx := context.TODO()
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metaV1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	customers := make([]DolittleCustomerTenant, 0)

	for _, ns := range namespaces.Items {
		nsName := ns.GetObjectMeta().GetName()

		if !strings.HasPrefix(nsName, "application-") {
			continue
		}

		annotationsMap := ns.GetObjectMeta().GetAnnotations()
		labelMap := ns.GetObjectMeta().GetLabels()

		applicationID := annotationsMap["dolittle.io/application-id"]
		tenantID := annotationsMap["dolittle.io/tenant-id"]

		tenant := DolittleTenant{
			Name: labelMap["tenant"],
			ID:   tenantID,
		}

		storageAccountInfo, _ := GetStorageAccountInfo(ctx, nsName, clientset)
		if !withSecrets {
			storageAccountInfo.Key = "XXX"
		}

		applications, _ := GetApplications(ctx, nsName, clientset)

		customers = append(customers, DolittleCustomerTenant{
			Namespace:           nsName,
			ApplicationID:       applicationID,
			Tenant:              tenant,
			AzureStorageAccount: storageAccountInfo,
			//IngressHosts:        hosts,
			Applications: applications,
		})
	}
	return customers
}

func GetStorageAccountInfo(ctx context.Context, namespace string, client *kubernetes.Clientset) (AzureStorageInfo, error) {
	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, "storage-account-secret", metaV1.GetOptions{})
	if err != nil {
		return AzureStorageInfo{}, err
	}
	return AzureStorageInfo{
		Name: string(secret.Data["azurestorageaccountname"]),
		Key:  string(secret.Data["azurestorageaccountkey"]),
	}, nil
}

//---------------------------------------------------------------------------------------------------------
//
//
//
//
//
// Customer specific
func GetIngressHosts(ctx context.Context, namespace string, client *kubernetes.Clientset, opts metaV1.ListOptions) ([]string, error) {
	hosts := make([]string, 0)
	obj, err := client.NetworkingV1().Ingresses(namespace).List(ctx, opts)
	if err != nil {
		return hosts, err
	}
	for _, item := range obj.Items {
		for _, rule := range item.Spec.Rules {
			if utils.StringArrayContains(hosts, rule.Host) {
				continue
			}
			hosts = append(hosts, rule.Host)
		}
	}
	return hosts, nil
}

// GetTenants Each application gets a configmap suffixed with "-tenants", so we mine this
func GetTenants(ctx context.Context, namespace string, client *kubernetes.Clientset, opts metaV1.ListOptions) ([]string, error) {
	tenants := make([]string, 0)
	configs, _ := client.CoreV1().ConfigMaps(namespace).List(ctx, opts)
	suffix := "-tenants"

	var found v1.ConfigMap

	for _, config := range configs.Items {
		if strings.HasSuffix(config.Name, suffix) {
			found = config
		}
	}

	if found.Name == "" {
		return tenants, errors.New("not-found")
	}

	rawTenants, ok := found.Data["tenants.json"]
	if !ok {
		return tenants, errors.New("not-found")

	}

	jsonObject := make(map[string]json.RawMessage)
	err := json.Unmarshal([]byte(rawTenants), &jsonObject)

	if err != nil {
		return tenants, err
	}

	for tenantID := range jsonObject {
		tenants = append(tenants, tenantID)
	}

	return tenants, nil
}

// GetMicroservices Check deployments for application label
func GetMicroservices(ctx context.Context, namespace string, client *kubernetes.Clientset, opts metaV1.ListOptions) ([]DolittleMicroservice, error) {
	microservices := make([]DolittleMicroservice, 0)
	deployments, err := client.AppsV1().Deployments(namespace).List(ctx, opts)
	if err != nil {
		return microservices, err
	}

	for _, deployment := range deployments.Items {
		name, ok := deployment.ObjectMeta.Labels["microservice"]
		if !ok {
			continue
		}

		application, _ := deployment.ObjectMeta.Labels["application"]
		environment, _ := deployment.ObjectMeta.Labels["environment"]
		microservices = append(microservices, DolittleMicroservice{
			ID:          deployment.ObjectMeta.Annotations["dolittle.io/microservice-id"],
			Name:        name,
			Application: application,
			Environment: environment,
		})
	}

	return microservices, nil
}

func GetShareName(ctx context.Context, namespace string, client *kubernetes.Clientset, opts metaV1.ListOptions) (string, error) {
	crons, err := client.BatchV1beta1().CronJobs(namespace).List(ctx, opts)
	if err != nil {
		return "", err
	}

	if len(crons.Items) == 0 {
		return "not-found", nil
	}
	return crons.Items[0].Spec.JobTemplate.Spec.Template.Spec.Volumes[0].AzureFile.ShareName, nil
}

func GetApplications(ctx context.Context, namespace string, client *kubernetes.Clientset) ([]DolittleApplication, error) {
	applications := make([]DolittleApplication, 0)
	obj, err := client.AppsV1().Deployments(namespace).List(ctx, metaV1.ListOptions{})
	if err != nil {
		return applications, err
	}

	applicationNames := make([]string, 0)
	for _, item := range obj.Items {
		tenantID, ok := item.ObjectMeta.Annotations["dolittle.io/tenant-id"]
		if !ok {
			continue
		}

		applicationID, ok := item.ObjectMeta.Annotations["dolittle.io/application-id"]
		if !ok {
			continue
		}

		applicationName, ok := item.ObjectMeta.Labels["application"]
		if !ok {
			continue
		}

		uniqueAppWithEnv := fmt.Sprintf("%s/%s", applicationName, item.ObjectMeta.Labels["environment"])
		if utils.StringArrayContains(applicationNames, uniqueAppWithEnv) {
			continue
		}

		applicationNames = append(applicationNames, uniqueAppWithEnv)
		applications = append(applications, DolittleApplication{
			ID:   applicationID,
			Name: applicationName,
			Tenant: DolittleTenant{
				Name: item.ObjectMeta.Labels["tenant"],
				ID:   tenantID,
			},
			Environment: item.ObjectMeta.Labels["environment"],
		})
	}

	for index, application := range applications {
		microservices, _ := GetMicroservices(ctx, namespace, client, metaV1.ListOptions{
			LabelSelector: fmt.Sprintf("tenant=%s,application=%s,environment=%s", application.Tenant.Name, application.Name, application.Environment),
		})
		applications[index].Microservices = microservices

		tenantIDS, _ := GetTenants(ctx, namespace, client, metaV1.ListOptions{
			LabelSelector: fmt.Sprintf("tenant=%s,application=%s,environment=%s", application.Tenant.Name, application.Name, application.Environment),
		})
		applications[index].TenantIDS = tenantIDS

		azureShareName, _ := GetShareName(ctx, namespace, client, metaV1.ListOptions{
			LabelSelector: fmt.Sprintf("tenant=%s,application=%s,environment=%s,infrastructure=Mongo", application.Tenant.Name, application.Name, application.Environment),
		})
		applications[index].AzureShareName = azureShareName

		ingressHosts, _ := GetIngressHosts(ctx, namespace, client, metaV1.ListOptions{
			LabelSelector: fmt.Sprintf("tenant=%s,application=%s,environment=%s", application.Tenant.Name, application.Name, application.Environment),
		})
		applications[index].IngressHosts = ingressHosts
	}

	return applications, nil
}
