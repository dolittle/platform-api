package microservice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	gitStorage "github.com/dolittle-entropy/platform-api/pkg/platform/storage/git"
	"github.com/itchyny/gojq"
	"github.com/spf13/cobra"
)

var buildCustomersCMD = &cobra.Command{
	Use:   "build-tenant-info",
	Short: "Write tenant info into the git repo",
	Long: `
	You need  the output from terraform

	terraform output -json
	go run main.go microservice build-tenant-info /Users/freshteapot/dolittle/git/Operations/Source/V3/Azure/azure.json
	`,
	Run: func(cmd *cobra.Command, args []string) {
		pathToFile := args[0]
		b, err := ioutil.ReadFile(pathToFile)
		if err != nil {
			fmt.Println(err)
			return
		}

		gitRepo := gitStorage.NewGitStorage(
			"git@github.com:freshteapot/test-deploy-key.git",
			"/tmp/dolittle-k8s",
			"auto-dev",
			// TODO fix this, then update deployment
			"/Users/freshteapot/dolittle/.ssh/test-deploy",
		)

		customers := extractCustomers(b)
		err = SaveCustomers(gitRepo, customers)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

func SaveCustomers(repo storage.Repo, customers []platform.TerraformCustomer) error {
	for _, customer := range customers {

		err := repo.SaveTenant(customer)
		if err != nil {
			return err
		}
	}
	return nil
}

func extractCustomers(data []byte) []platform.TerraformCustomer {
	var input interface{}
	json.Unmarshal(data, &input)

	// This would be easier if we added context in terraform
	jqQuery := `[.|to_entries | .[] | select(.value.value.azure_storage_account_name).value.value] | unique_by(.guid) | .[]`
	//query, err := gojq.Parse(`[.|to_entries | .[] | select(.value.value.customer).value.value.customer] | unique_by(.guid) | .[]`)
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
