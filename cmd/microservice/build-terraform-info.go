package microservice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	gitStorage "github.com/dolittle-entropy/platform-api/pkg/platform/storage/git"
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

	GIT_REPO_SSH_KEY="/Users/freshteapot/dolittle/.ssh/test-deploy" \
	GIT_REPO_BRANCH=auto-dev \
	GIT_REPO_URL="git@github.com:freshteapot/test-deploy-key.git" \
	go run main.go microservice build-terraform-info /Users/freshteapot/dolittle/git/Operations/Source/V3/Azure/azure.json
	`,
	Run: func(cmd *cobra.Command, args []string) {

		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logContext := logrus.StandardLogger()
		gitRepoConfig := initGit(logContext)

		platformEnvironment, _ := cmd.Flags().GetString("platform-environment")

		filterPlatformEnvironment := funk.ContainsString([]string{
			"dev",
			"prod",
		}, platformEnvironment)

		if !filterPlatformEnvironment {
			fmt.Println("The platform-environment can only be dev or prod")
			return
		}

		if (platformEnvironment == "dev") && (gitRepoConfig.Branch != platformEnvironment) {
			fmt.Println("The platform-environment does not match the branch")
			return
		}

		if (platformEnvironment == "prod") && (gitRepoConfig.Branch != "main") {
			fmt.Println("The platform-environment does not match the branch")
			return
		}

		pathToFile := args[0]
		b, err := ioutil.ReadFile(pathToFile)
		if err != nil {
			fmt.Println(err)
			return
		}

		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		customers := extractTerraformCustomers(platformEnvironment, b)

		err = saveTerraformCustomers(gitRepo, customers)
		if err != nil {
			fmt.Println(err)
			return
		}

		customerIDS := funk.Map(customers, func(customer platform.TerraformCustomer) string {
			return customer.GUID
		}).([]string)

		applications := extractTerraformApplications(customerIDS, b)
		err = saveTerraformApplications(gitRepo, applications)
		if err != nil {
			fmt.Println(err)
			return
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

func extractTerraformCustomers(platformEnvironment string, data []byte) []platform.TerraformCustomer {
	var input interface{}
	json.Unmarshal(data, &input)

	jqQuery := `[.|to_entries | .[] | select(.value.value.kind == "dolittle-customer" and .value.value.platform_environment == $platformEnvironment).value.value] | unique_by(.guid) | .[]`

	query, err := gojq.Parse(jqQuery)

	if err != nil {
		log.Fatalln(err)
	}

	code, err := gojq.Compile(
		query,
		gojq.WithVariables([]string{
			"$platformEnvironment",
		}),
	)

	if err != nil {
		log.Fatalln(err)
	}

	iter := code.Run(input, platformEnvironment)

	customers := make([]platform.TerraformCustomer, 0)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}

		if err, ok := v.(error); ok {
			log.Fatalln(err)
		}

		var c platform.TerraformCustomer
		b, _ := json.Marshal(v)
		err := json.Unmarshal(b, &c)
		if err != nil {
			log.Fatalln(err)
		}

		customers = append(customers, c)
	}

	return customers
}

func extractTerraformApplications(customerIDS []string, data []byte) []platform.TerraformApplication {
	var input interface{}
	json.Unmarshal(data, &input)

	jqQuery := `[.|to_entries | .[] | select(.value.value.kind == "dolittle-application" or .value.value.kind == "dolittle-application-with-resources").value.value] | unique_by(.guid) | .[]`
	query, err := gojq.Parse(jqQuery)

	if err != nil {
		log.Fatalln(err)
	}

	iter := query.Run(input)

	applications := make([]platform.TerraformApplication, 0)

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}

		if err, ok := v.(error); ok {
			log.Fatalln(err)
		}

		var a platform.TerraformApplication
		b, _ := json.Marshal(v)
		err := json.Unmarshal(b, &a)
		if err != nil {
			log.Fatalln(err)
		}

		if !funk.ContainsString(customerIDS, a.Customer.GUID) {
			fmt.Println(fmt.Sprintf("skipping as Customer %s (%s) is not on the list", a.Customer.Name, a.Customer.GUID))
			continue
		}
		applications = append(applications, a)
	}

	return applications
}

func init() {
	buildTerraformInfoCMD.Flags().String("platform-environment", "prod", "Platform environment (dev or prod), not linked to application environment")
}
