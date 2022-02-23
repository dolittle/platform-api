package update

import (
	"encoding/json"
	"errors"
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
	"github.com/spf13/viper"
	"github.com/thoas/go-funk"
)

var terraformCMD = &cobra.Command{
	Use:   "terraform",
	Short: "Update the terraform.json",
	Long: `
	You need the output from terraform

	cd Source/V3/Azure
	terraform output -json > azure.json

	GIT_REPO_BRANCH=dev \
	GIT_REPO_DRY_RUN=true \
	GIT_REPO_DIRECTORY="/tmp/dolittle-local-dev" \
	GIT_REPO_DIRECTORY_ONLY=true \
	go run main.go tools studio update terraform ~/Dolittle/Operations/Source/V3/Azure/azure.json
	`,
	Run: func(cmd *cobra.Command, args []string) {

		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logContext := logrus.StandardLogger()

		platformEnvironment := viper.GetString("tools.server.platformEnvironment")
		gitRepoConfig := git.InitGit(logContext, platformEnvironment)
		customerID, _ := cmd.Flags().GetString("customer-id")

		// TODO possibly change this if / when we introduce dynamic platform-environment
		filterPlatformEnvironment := funk.ContainsString([]string{
			"dev",
			"prod",
		}, platformEnvironment)

		if !filterPlatformEnvironment {
			logContext.Fatal("The platform-environment can only be dev or prod")
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

		if customerID != "" {
			filteredCustomer, err := findCustomer(customers, customerID)
			if err != nil {
				fmt.Println("Customer not found")
				return
			}
			customers = []platform.TerraformCustomer{
				filteredCustomer,
			}
		}

		err = saveTerraformCustomers(gitRepo, customers)
		if err != nil {
			logContext.WithField("error", err).Fatal("Failed to save terraform customers")
		}

		applications, err := extractTerraformApplications(customers, fileBytes)
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

func extractTerraformApplications(customers []platform.TerraformCustomer, data []byte) ([]platform.TerraformApplication, error) {
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

		customer, err := findCustomer(customers, a.Customer.GUID)
		if err != nil {
			fmt.Printf("skipping as Customer (%s) is not found\n", a.Customer.GUID)
			continue
		}
		// We map to the actual customer, as the customer struct has all the customer values
		a.Customer = customer
		applications = append(applications, a)
	}

	return applications, nil
}

func findCustomer(customers []platform.TerraformCustomer, customerID string) (platform.TerraformCustomer, error) {
	for _, customer := range customers {
		if customer.GUID == customerID {
			return customer, nil
		}
	}
	return platform.TerraformCustomer{}, errors.New("not.found")
}

func init() {
	terraformCMD.Flags().String("customer-id", "", "Specific customer-id to use (optional)")
}
