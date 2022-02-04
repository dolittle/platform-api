package manual

import (
	"context"

	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/automate"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Repo struct {
	client      kubernetes.Interface
	k8sRepoV2   k8s.Repo
	storageRepo storage.Repo
	logContext  logrus.FieldLogger
}

func NewManualHelper(
	client kubernetes.Interface,
	k8sRepoV2 k8s.Repo,
	storageRepo storage.Repo,
	logContext logrus.FieldLogger,
) Repo {
	return Repo{
		client:      client,
		k8sRepoV2:   k8sRepoV2,
		storageRepo: storageRepo,
		logContext:  logContext,
	}
}

func (r Repo) GatherOne(platformEnvironment string, namespace string) (storage.JSONApplication, error) {
	application := storage.JSONApplication{
		Environments: make([]storage.JSONEnvironment, 0),
	}
	ctx := context.TODO()
	client := r.client

	namespaceResource, err := client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})

	if err != nil {
		return storage.JSONApplication{}, err
	}
	// Confirm it has applicaiton terraform
	application.ID = namespaceResource.Annotations["dolittle.io/application-id"]
	application.Name = namespaceResource.Labels["application"]
	application.TenantID = namespaceResource.Annotations["dolittle.io/tenant-id"]
	application.TenantName = namespaceResource.Labels["tenant"]

	//Get customerTenants
	// TODO write this to storage?
	customerTenants := r.GetCustomerTenants(ctx, namespace)
	environmentNames := r.GetEnvironmentNames(ctx, namespace)

	for _, environmentName := range environmentNames {
		environment := storage.JSONEnvironment{
			Name: environmentName,
			CustomerTenants: funk.Filter(customerTenants, func(customerTenant platform.CustomerTenantInfo) bool {
				return customerTenant.Environment == environmentName
			}).([]platform.CustomerTenantInfo),
			WelcomeMicroserviceID: "",
		}
		application.Environments = append(application.Environments, environment)
	}

	return application, nil
}

func (r Repo) GetEnvironmentNames(ctx context.Context, namespace string) []string {
	client := r.client
	customerTenantsConfigMaps, err := automate.GetCustomerTenantsConfigMaps(ctx, client, namespace)
	if err != nil {
		panic(err)
	}

	environments := make([]string, 0)
	for _, configMap := range customerTenantsConfigMaps {
		environments = append(environments, configMap.Labels["environment"])
	}
	return environments
}
