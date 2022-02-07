package automate

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
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
		platformEnvironment, _ := cmd.Flags().GetString("platform-environment")
		gitRepoConfig := git.InitGit(logContext, platformEnvironment)

		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		customerID, _ := cmd.Flags().GetString("customer-id")
		environment, _ := cmd.Flags().GetString("environment")
		applicationID, _ := cmd.Flags().GetString("application-id")

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

		client, config := platformK8s.InitKubernetesClient()

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

		// How does it know what todo with the microservice?
		// Look up the id from the customer Tenants in each environment?
		// How do I give it the signal?

		// This is used to make local dev happy
		// TODO FIX
		isProduction := false
		welcomeImage := welcome.Image
		k8sDolittleRepo := platformK8s.NewK8sRepo(client, config, logContext.WithField("context", "k8s-repo"))
		simpleRepo := k8sSimple.NewSimpleRepo(platformEnvironment, client, k8sDolittleRepo)
		// TODO refactor when it works
		err = platformApplication.CreateApplicationAndEnvironmentAndWelcomeMicroservice(
			client,
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
			// TODO failed to update state
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
	createApplicationCMD.Flags().String("platform-environment", "dev", "Platform environment (dev or prod), not linked to application environment")
}
