package application

import (
	"fmt"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/application/k8s"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/simple"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/sirupsen/logrus"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateDevelopmentBinding() error {
	return nil
}

// TODO at some point, we might want to explode this into environments
func CreateApplicationAndEnvironmentAndWelcomeMicroservice(
	client kubernetes.Interface,
	storageRepo storage.RepoMicroservice,
	simplRepo simple.Repo,
	k8sDolittleRepo platformK8s.K8sRepo,
	application storage.JSONApplication2,
	terraformCustomer platform.TerraformCustomer,
	terraformApplication platform.TerraformApplication,
	isProduction bool,
	logContext logrus.FieldLogger,
) error {
	dockerconfigjson := k8s.MakeCustomerAcrDockerConfig(terraformCustomer)
	// TODO this comes from creating a resource_group and
	azureGroupId := terraformApplication.GroupID
	azureStorageAccountName := terraformCustomer.AzureStorageAccountName
	// TODO Get key from terraform, write to studio
	azureStorageAccountKey := terraformCustomer.AzureStorageAccountKey
	// TODO this should be a config setting or a env variable
	welcomeImage := "nginxdemos/hello:latest"

	tenantInfo := dolittleK8s.Tenant{
		Name: application.TenantName,
		ID:   application.TenantID,
	}

	applicationInfo := dolittleK8s.Application{
		Name: application.Name,
		ID:   application.ID,
	}

	welcomeMicroservice := platform.HttpInputSimpleInfo{
		MicroserviceBase: platform.MicroserviceBase{
			Dolittle: platform.HttpInputDolittle{
				ApplicationID:  application.ID,
				TenantID:       application.TenantID,
				MicroserviceID: "",
			},
			Name: "Welcome",
			Kind: platform.MicroserviceKindSimple,
		},
		Extra: platform.HttpInputSimpleExtra{
			Headimage:    welcomeImage,
			Runtimeimage: "none",
			Ingress: platform.HttpInputSimpleIngress{
				Host:     "",
				Path:     "/welcome-to-dolittle",
				Pathtype: string(networkingv1.PathTypePrefix),
			},
		},
	}

	r := k8s.Resources{}
	r.ServiceAccounts = k8s.NewServiceAccountsInfo(tenantInfo, applicationInfo)

	r.Namespace = k8s.NewNamespace(tenantInfo, applicationInfo)
	r.Acr = k8s.NewAcr(tenantInfo, applicationInfo, dockerconfigjson)
	r.Storage = k8s.NewStorage(tenantInfo, applicationInfo, azureStorageAccountName, azureStorageAccountKey)

	// to add to rbac
	r.Rbac = k8s.NewDeveloperRole(tenantInfo, applicationInfo, azureGroupId)

	// TODO Create service account for devops
	// TODO Joel is on the case (figure out how to add it or call it after, as it we might want to decouple it from here)

	// Create rbac
	// Create environments
	for _, environment := range application.Environments {
		mongoSettings := k8s.MongoSettings{
			ShareName:       azureStorageAccountName,
			CronJobSchedule: fmt.Sprintf("%d * * * *", GetRandomMinutes()),
			VolumeSize:      "8Gi",
		}

		environmentResource := k8s.NewEnvironment(environment.Name, tenantInfo, applicationInfo, mongoSettings, environment.CustomerTenants)
		r.Environments = append(r.Environments, environmentResource)
	}

	// TODO figure out how to know if we are local dev
	// Add Local dev bindings
	if !isProduction {
		r.LocalDevRoleBindingToDeveloper = k8s.NewLocalDevRoleBindingToDeveloper(tenantInfo, applicationInfo)
	}
	// Create

	err := k8s.Do(client, r, k8sDolittleRepo)
	if err != nil {
		return err
	}

	// Create welcome microservice
	namesapce := r.Namespace.Name
	for _, environment := range application.Environments {
		customerTenants := environment.CustomerTenants

		microservice := welcomeMicroservice
		microservice.Dolittle.MicroserviceID = customerTenants[0].MicroservicesRel[0].MicroserviceID
		microservice.Environment = environment.Name

		// TODO Would be nice to hoist this to the creation of the application, so this is semi immutable
		err = storageRepo.SaveMicroservice(
			microservice.Dolittle.TenantID,
			microservice.Dolittle.ApplicationID,
			microservice.Environment,
			microservice.Dolittle.MicroserviceID,
			microservice,
		)

		if err != nil {
			return err
		}

		err := simplRepo.Create(namesapce, tenantInfo, applicationInfo, customerTenants, microservice)
		if err != nil {
			return err
		}
	}

	return nil
}
