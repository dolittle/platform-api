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
)

var buildTerraformInfoCMD = &cobra.Command{
	Use:   "build-terraform-info",
	Short: "Write tenant info into the git repo",
	Long: `
	You need  the output from terraform

	terraform output -json
	GIT_BRANCH="auto-dev" \
	go run main.go microservice build-terraform-info /Users/freshteapot/dolittle/git/Operations/Source/V3/Azure/azure.json
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logContext := logrus.StandardLogger()
		gitRepoConfig := initGit(logContext)

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

		customers := extractTerraformCustomers(b)
		err = saveTerraformCustomers(gitRepo, customers)
		if err != nil {
			fmt.Println(err)
			return
		}

		applications := extractTerraformApplications(b)
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

func extractTerraformCustomers(data []byte) []platform.TerraformCustomer {
	var input interface{}
	json.Unmarshal(data, &input)

	jqQuery := `[.|to_entries | .[] | select(.value.value.kind=="dolittle-customer").value.value] | unique_by(.guid) | .[]`
	query, err := gojq.Parse(jqQuery)

	if err != nil {
		log.Fatalln(err)
	}

	iter := query.Run(input) // or query.RunWithContext

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

func extractTerraformApplications(data []byte) []platform.TerraformApplication {
	var input interface{}
	json.Unmarshal(data, &input)

	jqQuery := `[.|to_entries | .[] | select(.value.value.kind=="dolittle-application").value.value] | unique_by(.guid) | .[]`
	query, err := gojq.Parse(jqQuery)

	if err != nil {
		log.Fatalln(err)
	}

	iter := query.Run(input) // or query.RunWithContext

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

		applications = append(applications, a)
	}

	return applications
}
