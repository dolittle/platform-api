package create

import (
	"fmt"
	"os"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/git"
	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/private"
	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var privateCMD = &cobra.Command{
	Use:   "private",
	Short: "Create a private microservice in kubernetes",
	Long: `
Create private microservices without an ingress or a domain.

APPLICATION_ID=df2011f9-9407-447e-8b64-df06d6aaf9a1 \
APPLICATION_NAME=Test \
MICROSERVICE_NAME=TestMicroservice \
ENVIRONMENT=dev \
TENANT_NAME="Test tenant" \
TENANT_ID=a10d966b-ded9-4281-88b8-27a4123bc716 \
CUSTOMER_ID=1d258a46-dab6-4ad8-b148-21b324f2c40f \
HEAD_IMAGE=nginxdemos/hello \
RUNTIME_IMAGE=dolittle/runtime:7.8.0 \
go run main.go tools microservice create private
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)
		logContext := logrus.StandardLogger()

		// @joel maybe we don't allow the user to set their own microserviceID. But it makes importing
		// already existing microservices in another cluster harder as they already have an established microserviceID
		// but we also then should check for conflicts with existing microserviceIDs
		microserviceID := viper.GetString("tools.microservice.microserviceId")

		applicationName := viper.GetString("tools.microservice.create.application.name")
		if applicationName == "" {
			logContext.Fatal("no application name defined")
		}
		applicationID := viper.GetString("tools.microservice.applicationId")
		if applicationID == "" {
			logContext.Fatal("no application ID defined")
		}
		application := dolittleK8s.Application{
			Name: applicationName,
			ID:   applicationID,
		}
		namespace := fmt.Sprintf("application-%s", applicationID)

		environment := viper.GetString("tools.microservice.environment")
		if environment == "" {
			logContext.Fatal("no environment defined")
		}

		// @joel this could be dynamic with somethign like --tenant-name-1 or something idk
		tenantName := viper.GetString("tools.microservice.create.tenant.name")
		if tenantName == "" {
			logContext.Fatal("no tenant name defined")
		}

		tenantID := viper.GetString("tools.microservice.create.tenant.id")

		tenant := dolittleK8s.Tenant{
			Name: tenantName,
			ID:   tenantID,
		}

		tenantInfo := platform.CustomerTenantInfo{
			Alias:            tenantName,
			Environment:      environment,
			CustomerTenantID: tenantID,
			// TODO it's a private microservice so we shouldn't care about this, maybe make another type
			Hosts: []platform.CustomerTenantHost{},
			MicroservicesRel: []platform.CustomerTenantMicroserviceRel{
				{
					MicroserviceID: microserviceID,
					Hash:           dolittleK8s.ResourcePrefix(microserviceID, tenantID),
				},
			},
		}
		customerTenants := []platform.CustomerTenantInfo{tenantInfo}

		customerID := viper.GetString("tools.microservice.create.customerID")
		if customerID == "" {
			logContext.Fatal("no customer ID defined")
		}

		microserviceName := viper.GetString("tools.microservice.create.microservice.name")
		if microserviceName == "" {
			logContext.Fatal("no microservice name defined")
		}

		kind := platform.MicroserviceKindPrivate

		headImage := viper.GetString("tools.microservice.create.headImage")
		if headImage == "" {
			logContext.Fatal("no head image defined")
		}
		runtimeImage := viper.GetString("tools.microservice.create.runtimeImage")
		if runtimeImage == "" {
			logContext.Fatal("no runtime image defined")
		}

		input := platform.HttpInputSimpleInfo{
			MicroserviceBase: platform.MicroserviceBase{
				Dolittle: platform.HttpInputDolittle{
					ApplicationID:  applicationID,
					CustomerID:     customerID,
					MicroserviceID: microserviceID,
				},
				Name:        microserviceName,
				Kind:        kind,
				Environment: environment,
			},
			Extra: platform.HttpInputSimpleExtra{
				Headimage:    headImage,
				Runtimeimage: runtimeImage,
				// also empty as we ain't gonna use it
				Ingress: platform.HttpInputSimpleIngress{},
			},
		}

		k8sClient, k8sConfig := platformK8s.InitKubernetesClient()
		k8sRepo := platformK8s.NewK8sRepo(k8sClient, k8sConfig, logContext.WithField("context", "k8s-repo"))
		k8sRepoV2 := k8s.NewRepo(k8sClient, logContext.WithField("context", "k8s-repo-v2"))

		repo := private.NewPrivateRepo(k8sClient, k8sRepo, k8sRepoV2)

		platformEnvironment := viper.GetString("tools.server.platformEnvironment")
		gitRepoConfig := git.InitGit(logContext, platformEnvironment)

		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		err := repo.Create(namespace, tenant, application, customerTenants, input)
		if err != nil {
			logContext.Panic(err)
		}

		err = gitRepo.SaveMicroservice(customerID, applicationID, environment, microserviceID, input)
		if err != nil {
			logContext.Panic(err)
		}
	},
}
