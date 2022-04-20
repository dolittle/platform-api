package application

import (
	"fmt"

	"github.com/dolittle/platform-api/pkg/azure"
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

func CreateApplicationAndEnvironmentAndWelcomeMicroservice(
	client kubernetes.Interface,
	storageRepo storage.RepoMicroservice,
	simpleRepo simple.Repo,
	k8sDolittleRepo platformK8s.K8sRepo,
	application storage.JSONApplication,
	terraformCustomer platform.TerraformCustomer,
	terraformApplication platform.TerraformApplication,
	isProduction bool,
	welcomeImage string,
	logContext logrus.FieldLogger,
) error {
	dockerconfigjson := k8s.MakeCustomerAcrDockerConfig(terraformCustomer)
	azureGroupId := terraformApplication.GroupID
	azureStorageAccountName := terraformCustomer.AzureStorageAccountName
	azureStorageAccountKey := terraformCustomer.AzureStorageAccountKey

	tenantInfo := dolittleK8s.Tenant{
		Name: application.CustomerName,
		ID:   application.CustomerID,
	}

	applicationInfo := dolittleK8s.Application{
		Name: application.Name,
		ID:   application.ID,
	}

	welcomeMicroservice := platform.HttpInputSimpleInfo{
		MicroserviceBase: platform.MicroserviceBase{
			Dolittle: platform.HttpInputDolittle{
				ApplicationID:  application.ID,
				CustomerID:     application.CustomerID,
				MicroserviceID: "",
			},
			Name: "Welcome",
			Kind: platform.MicroserviceKindSimple,
		},
		Extra: platform.HttpInputSimpleExtra{
			Headimage:    welcomeImage,
			Runtimeimage: "none",
			Ingress: platform.HttpInputSimpleIngress{
				Path:     "/welcome-to-dolittle",
				Pathtype: string(networkingv1.PathTypePrefix),
			},
			Ispublic: true,
		},
	}

	r := k8s.Resources{}
	r.ServiceAccounts = k8s.NewServiceAccountsInfo(tenantInfo, applicationInfo)

	r.Namespace = k8s.NewNamespace(tenantInfo, applicationInfo)
	r.Acr = k8s.NewAcr(tenantInfo, applicationInfo, dockerconfigjson)
	r.Storage = k8s.NewStorage(tenantInfo, applicationInfo, azureStorageAccountName, azureStorageAccountKey)

	r.DeveloperRbac = k8s.NewDeveloperRbac(tenantInfo, applicationInfo, azureGroupId)

	// Create rbac
	// Create environments
	for _, environment := range application.Environments {
		shareName := azure.CreateBackupFileShareName(application.Name, environment.Name)

		mongoSettings := k8s.MongoSettings{
			ShareName:       shareName,
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
	namespace := r.Namespace.Name
	for _, environment := range application.Environments {
		customerTenants := environment.CustomerTenants

		skipMicroserviceCreation := len(customerTenants) == 0
		if skipMicroserviceCreation {
			continue
		}

		microservice := welcomeMicroservice
		microservice.Dolittle.MicroserviceID = customerTenants[0].MicroservicesRel[0].MicroserviceID
		microservice.Environment = environment.Name

		// TODO Would be nice to hoist this to the creation of the application, so this is semi immutable
		err = storageRepo.SaveMicroservice(
			microservice.Dolittle.CustomerID,
			microservice.Dolittle.ApplicationID,
			microservice.Environment,
			microservice.Dolittle.MicroserviceID,
			microservice,
		)

		if err != nil {
			return err
		}

		err := simpleRepo.Create(namespace, tenantInfo, applicationInfo, customerTenants, microservice)
		if err != nil {
			return err
		}
	}

	return nil
}
