package automate

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dolittle/platform-api/pkg/azure"
	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/k8s"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"

	k8sSimple "github.com/dolittle/platform-api/pkg/platform/microservice/simple/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/welcome"

	"github.com/dolittle/platform-api/pkg/git"
	platformApplication "github.com/dolittle/platform-api/pkg/platform/application"
)

var createApplicationCMD = &cobra.Command{
	Use:   "create-application",
	Short: "Create Application in kubernetes",
	Long: `
In kubernetes, create application
	--with-environments 
		include environments
	
	--with-welcome-microservice
		include welcome microservice per environment (requires with-environments)
	
	--customer-id=XXX (TODO)
		Customer ID of where the tenant lives
	--environment=XXX (TODO)
		If environment is included only this
	
	--platform-environment=dev
		This is linked to the data in terraform, to signify what type of customer this is

	--is-production=true
		Signal that we are in production mode

	go run main.go tools automate create-application \
	--with-environments \
	--with-welcome-microservice \
	--application-id=XXX \
	
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// Make sure we use git variables
		git.SetupViper()
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logContext := logrus.StandardLogger()
		platformEnvironment := viper.GetString("tools.server.platformEnvironment")

		gitRepoConfig := git.InitGit(logContext, platformEnvironment)

		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		customerID, _ := cmd.Flags().GetString("customer-id")
		environment, _ := cmd.Flags().GetString("environment")
		applicationID, _ := cmd.Flags().GetString("application-id")

		isProduction, _ := cmd.Flags().GetBool("is-production")
		withEnvironments, _ := cmd.Flags().GetBool("with-environments")

		withWelcomeMicroservice, _ := cmd.Flags().GetBool("with-welcome-microservice")

		if customerID == "" {
			fmt.Println("An --customer-id  is required")
			return
		}

		if applicationID == "" {
			fmt.Println("An --application-id  is required")
			return
		}

		if !withEnvironments && environment == "" {
			fmt.Println("An --environment is required when not using --withEnvironments")
			return
		}

		if !withEnvironments {
			fmt.Println("TODO: create application without environments")
			return
		}

		if !withWelcomeMicroservice {
			fmt.Println("TODO: create application with environments but without welcome microservice")
			return
		}

		k8sClient, k8sConfig := platformK8s.InitKubernetesClient()

		terraformCustomer, err := gitRepo.GetTerraformTenant(customerID)
		if err != nil {
			panic(err.Error())

		}

		tenant := dolittleK8s.Tenant{
			ID:   terraformCustomer.GUID,
			Name: terraformCustomer.Name,
		}

		terraformApplication, err := gitRepo.GetTerraformApplication(tenant.ID, applicationID)
		if err != nil {
			panic(err.Error())
		}

		application, err := gitRepo.GetApplication(customerID, terraformApplication.ApplicationID)
		if err != nil {
			panic(err.Error())
		}

		welcomeImage := welcome.Image
		k8sDolittleRepo := platformK8s.NewK8sRepo(k8sClient, k8sConfig, logContext.WithField("context", "k8s-repo"))
		k8sRepoV2 := k8s.NewRepo(k8sClient, logContext.WithField("context", "k8s-repo-v2"))
		simpleRepo := k8sSimple.NewSimpleRepo(k8sClient, k8sDolittleRepo, k8sRepoV2, isProduction)

		azureStorageAccountName := terraformCustomer.AzureStorageAccountName
		azureStorageAccountKey := terraformCustomer.AzureStorageAccountKey
		// ensure that the fileshares exist before creating the k8s resources
		for _, environment := range application.Environments {
			fileShare := azure.CreateBackupFileShareName(application.Name, environment.Name)

			err := azure.EnsureFileShareExists(azureStorageAccountName, azureStorageAccountKey, fileShare)
			if err != nil {
				panic(err.Error())
			}
		}

		err = platformApplication.CreateApplicationAndEnvironmentAndWelcomeMicroservice(
			k8sClient,
			gitRepo,
			simpleRepo,
			k8sDolittleRepo,
			application,
			terraformCustomer,
			terraformApplication,
			isProduction,
			welcomeImage,
			logContext,
		)

		if err != nil {
			panic(err.Error())
		}

		application.Status.State = storage.BuildStatusStateFinishedSuccess
		application.Status.FinishedAt = time.Now().UTC().Format(time.RFC3339)

		err = gitRepo.SaveApplication(application)
		if err != nil {
			panic(err.Error())
		}
	},
}

func init() {
	createApplicationCMD.Flags().String("customer-id", "", "Customer ID of where the application lives")
	createApplicationCMD.Flags().String("environment", "", "Environment to use")
	createApplicationCMD.Flags().String("application-id", "", "Application ID to use")
	createApplicationCMD.Flags().Bool("with-environments", false, "Setup environments")
	createApplicationCMD.Flags().Bool("with-welcome-microservice", false, "Add welcome microservice to each environment")
	createApplicationCMD.Flags().Bool("is-production", false, "Signal this is in production mode")
}
