package tools

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
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

	go run main.go tools create-customer-hcl --name="Test Marka" --platform-environment="dev"

	`,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			fmt.Println("A customer name is required")
			return
		}

		source, _ := cmd.Flags().GetString("source")
		platformEnvironment, _ := cmd.Flags().GetString("platform-environment")
		guid := uuid.New().String()

		// Lowercase and replace spaces
		suffix := strings.ToLower(name)
		suffix = strings.ReplaceAll(suffix, " ", "_")
		tfName := fmt.Sprintf("customer_%s", suffix)

		customer := tfCustomer{
			Name:                name,
			Source:              source,
			PlatformEnvironment: platformEnvironment,
			Guid:                guid,
			TfName:              tfName,
		}

		f := hclwrite.NewEmptyFile()
		generateTerraformForCustomer(f.Body(), customer)

		fmt.Printf("%s", f.Bytes())
	},
}

func generateTerraformForCustomer(root *hclwrite.Body, customer tfCustomer) {
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
	createCustomerHclCMD.Flags().String("platform-environment", "prod", "Platform environment (dev or prod), not linked to application environment")
	createCustomerHclCMD.Flags().String("source", "./modules/dolittle-customer-with-resources", "Override with specific source of the module")
}
