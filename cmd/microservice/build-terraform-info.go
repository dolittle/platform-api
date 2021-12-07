package microservice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/dolittle/platform-api/pkg/git"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"
	"github.com/itchyny/gojq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/thoas/go-funk"
)

var buildTerraformInfoCMD = &cobra.Command{
	Use:   "build-terraform-info",
	Short: "Write tenant info into the git repo",
	Long: `
	You need  the output from terraform

	cd Source/V3/Azure
	terraform output -json > azure.json

	GIT_REPO_BRANCH=dev \
	GIT_REPO_DRY_RUN=true \
	GIT_REPO_DIRECTORY="/tmp/dolittle-local-dev" \
	GIT_REPO_DIRECTORY_ONLY=true \
	go run main.go microservice build-terraform-info ~/Dolittle/Operations/Source/V3/Azure/azure.json

	GIT_REPO_SSH_KEY="/Users/freshteapot/dolittle/.ssh/test-deploy" \
	GIT_REPO_BRANCH=dev \
	GIT_REPO_DRY_RUN=true \
	GIT_REPO_URL="git@github.com:freshteapot/test-deploy-key.git" \
	go run main.go microservice build-terraform-info /Users/freshteapot/dolittle/git/Operations/Source/V3/Azure/azure.json
	`,
	Run: func(cmd *cobra.Command, args []string) {

		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logContext := logrus.StandardLogger()
		gitRepoConfig := git.InitGit(logContext)

		platformEnvironment, _ := cmd.Flags().GetString("platform-environment")

		filterPlatformEnvironment := funk.ContainsString([]string{
			"dev",
			"prod",
		}, platformEnvironment)

		if !filterPlatformEnvironment {
			logContext.Fatal("The platform-environment can only be dev or prod")
		}

		if (platformEnvironment == "dev") && (gitRepoConfig.Branch != platformEnvironment) {
			logContext.Fatal("The platform-environment does not match the branch")
		}

		if (platformEnvironment == "prod") && (gitRepoConfig.Branch != "main") {
			logContext.Fatal("The platform-environment does not match the branch")
		}

		pathToFile := args[0]
		fileBytes, err := ioutil.ReadFile(pathToFile)
		if err != nil {
			logContext.WithField("error", err).Fatal("Failed to find path")
		}

		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		customers, err := extractTerraformCustomers(platformEnvironment, fileBytes)
		if err != nil {
			logContext.WithField("error", err).Fatal("Failed to extract terraform customers")
		}

		err = saveTerraformCustomers(gitRepo, customers)
		if err != nil {
			logContext.WithField("error", err).Fatal("Failed to save terraform customers")
		}

		customerIDS := funk.Map(customers, func(customer platform.TerraformCustomer) string {
			return customer.GUID
		}).([]string)

		applications, err := extractTerraformApplications(customerIDS, fileBytes)
		if err != nil {
			logContext.WithField("error", err).Fatal("Failed to extract terraform applications")
		}

		err = saveTerraformApplications(gitRepo, applications)
		if err != nil {
			logContext.WithField("error", err).Fatal("Failed to save terraform applications")
		}
	},
}

func saveTerraformCustomers(repo storage.Repo, customers []platform.TerraformCustomer) error {
	for _, customer := range customers {

		err := repo.SaveTerraformTenant(customer)
		if err != nil {
			return err
		}
	}
	return nil
}

func saveTerraformApplications(repo storage.Repo, applications []platform.TerraformApplication) error {
	for _, application := range applications {

		err := repo.SaveTerraformApplication(application)
		if err != nil {
			return err
		}
	}
	return nil
}

func extractTerraformCustomers(platformEnvironment string, data []byte) ([]platform.TerraformCustomer, error) {
	var input interface{}
	customers := make([]platform.TerraformCustomer, 0)
	json.Unmarshal(data, &input)

	jqQuery := `[.|to_entries | .[] | select(.value.value.kind == "dolittle-customer" and .value.value.platform_environment == $platformEnvironment).value.value] | unique_by(.guid) | .[]`

	query, err := gojq.Parse(jqQuery)

	if err != nil {
		return customers, err
	}

	code, err := gojq.Compile(
		query,
		gojq.WithVariables([]string{
			"$platformEnvironment",
		}),
	)

	if err != nil {
		return customers, err
	}

	iter := code.Run(input, platformEnvironment)

	for {
		value, ok := iter.Next()
		if !ok {
			break
		}

		if err, ok := value.(error); ok {
			return customers, err
		}

		var terraformCustomer platform.TerraformCustomer
		valueBytes, _ := json.Marshal(value)
		err := json.Unmarshal(valueBytes, &terraformCustomer)
		if err != nil {
			return customers, err
		}

		customers = append(customers, terraformCustomer)
	}

	return customers, nil
}

func extractTerraformApplications(customerIDS []string, data []byte) ([]platform.TerraformApplication, error) {
	var input interface{}
	applications := make([]platform.TerraformApplication, 0)

	json.Unmarshal(data, &input)

	jqQuery := `[.|to_entries | .[] | select(.value.value.kind == "dolittle-application" or .value.value.kind == "dolittle-application-with-resources").value.value] | unique_by(.guid) | .[]`
	query, err := gojq.Parse(jqQuery)

	if err != nil {
		return applications, err
	}

	iter := query.Run(input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}

		if err, ok := v.(error); ok {
			return applications, err
		}

		var a platform.TerraformApplication
		b, _ := json.Marshal(v)
		err := json.Unmarshal(b, &a)
		if err != nil {
			return applications, err
		}

		if !funk.ContainsString(customerIDS, a.Customer.GUID) {
			fmt.Println(fmt.Sprintf("skipping as Customer %s (%s) is not on the list", a.Customer.Name, a.Customer.GUID))
			continue
		}
		applications = append(applications, a)
	}

	return applications, nil
}

func init() {
	buildTerraformInfoCMD.Flags().String("platform-environment", "prod", "Platform environment (dev or prod), not linked to application environment")
}
