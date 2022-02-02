package terraform

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zclconf/go-cty/cty"
)

type tfCustomer struct {
	Source              string
	Guid                string
	Name                string
	TfName              string
	PlatformEnvironment string
}

var createCustomerHclCMD = &cobra.Command{
	Use:   "create-customer-hcl",
	Short: "Write terraform output for a customer",
	Long: `
	Outputs a new Dolittle platform customer in hcl to stdout.

	go run main.go tools terraform create-customer-hcl --name="Test Marka" --platform-environment="dev"
	terraform init
	terraform plan -target="module.customer_1266827b-51e6-4be6-ae18-a0ea4638c2ab"

	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logger := logrus.StandardLogger()

		dryRun, _ := cmd.Flags().GetBool("dry-run")

		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			fmt.Println("A customer name is required")
			return
		}

		customerID, _ := cmd.Flags().GetString("id")
		if customerID == "" {
			customerID = uuid.New().String()
		}

		source, _ := cmd.Flags().GetString("source")
		platformEnvironment, _ := cmd.Flags().GetString("platform-environment")

		rootFolder := ""
		if !dryRun {
			if len(args) == 0 {
				logger.Error("Specify the rootFolder to write to")
				return
			}
			rootFolder = args[0]
		}

		// TODO changing from name to UUID
		// Lowercase and replace spaces
		//suffix := strings.ToLower(name)
		//suffix = strings.ReplaceAll(suffix, " ", "_")
		tfName := fmt.Sprintf("customer_%s", customerID)

		customer := tfCustomer{
			Name:                name,
			Source:              source,
			PlatformEnvironment: platformEnvironment,
			Guid:                customerID,
			TfName:              tfName,
		}

		f := hclwrite.NewEmptyFile()
		generateTerraformForCustomer(f.Body(), customer)

		if dryRun {
			fmt.Printf("%s", f.Bytes())
			return
		}

		suffix := fmt.Sprintf("%s.tf", tfName)
		err := writeToTerraform(rootFolder, suffix, f.Bytes())
		if err != nil {
			logger.WithFields(logrus.Fields{
				"error": err,
			}).Error("Failed to write to terraform")
			return
		}

		fmt.Printf(`
cd %s
terraform init
terraform plan -target="module.%s"
`,
			getTerraformDirectory(rootFolder),
			tfName,
		)
	},
}

func generateTerraformForCustomer(root *hclwrite.Body, customer tfCustomer) {
	// TODO moving to uuid at the end
	// TODO need to migrate? or do we?
	/*
		module "customer_test_marka" {
			module_name          = "customer_test_marka"
			source               = "./modules/dolittle-customer-with-resources"
			guid                 = "XXX"
			name                 = "Test Marka"
			platform_environment = "dev"
		}

		output "customer_test_marka" {
			value = module.customer_test_marka
		}
	*/

	moduleBlock := root.AppendNewBlock("module", []string{customer.TfName})
	moduleBody := moduleBlock.Body()
	moduleBody.SetAttributeValue("module_name", cty.StringVal(customer.TfName))
	moduleBody.SetAttributeValue("source", cty.StringVal(customer.Source))
	moduleBody.SetAttributeValue("guid", cty.StringVal(customer.Guid))
	moduleBody.SetAttributeValue("name", cty.StringVal(customer.Name))
	moduleBody.SetAttributeValue("platform_environment", cty.StringVal(customer.PlatformEnvironment))

	root.AppendNewline()

	body := root.AppendNewBlock("output", []string{customer.TfName})
	body.Body().SetAttributeTraversal("value", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "module",
		},
		hcl.TraverseAttr{
			Name: customer.TfName,
		},
	})

	body.Body().SetAttributeValue("sensitive", cty.True)
}

func init() {
	createCustomerHclCMD.Flags().String("name", "", "Name of customer (readable)")
	createCustomerHclCMD.Flags().String("id", "", "Specific customer id to use")
	createCustomerHclCMD.Flags().String("platform-environment", "dev", "Platform environment (dev or prod), not linked to application environment")
	createCustomerHclCMD.Flags().String("source", "./modules/dolittle-customer-with-resources", "Override with specific source of the module")
	createCustomerHclCMD.Flags().Bool("dry-run", false, "Will not write to disk")

}

func writeToTerraform(rootFolder string, suffix string, data []byte) error {

	// TODO Do we want structure in terraform?
	tfDirectory := getTerraformDirectory(rootFolder)

	file, err := os.Create(filepath.Join(tfDirectory, suffix))
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.Write(data)
	return err
}
func getTerraformDirectory(rootFolder string) string {

	return filepath.Join(rootFolder,
		"Source",
		"V3",
		"Azure",
	)
}
