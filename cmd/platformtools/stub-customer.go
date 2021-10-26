package platformtools

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/spf13/cobra"
	"github.com/zclconf/go-cty/cty"
)

var stubCustomerCMD = &cobra.Command{
	Use:   "stub-customer",
	Short: "Write terraform output for a customer",
	Long: `

	`,
	Run: func(cmd *cobra.Command, args []string) {
		/*
		   module "customer_test_marka" {
		     source               = "./modules/dolittle-customer-with-resources"
		     guid                 = "XXX"
		     name                 = "Test Marka"
		     platform_environment = "dev"
		   }

		   output "customer_test_marka" {
		     value = module.customer_test_marka
		   }
		*/

		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			fmt.Println("A customer name is requried")
			return
		}

		source, _ := cmd.Flags().GetString("source")
		platformEnvironment, _ := cmd.Flags().GetString("platform-environment")
		guid := uuid.New().String()

		// Lowercase and replace spaces
		suffix := strings.ToLower(name)
		suffix = strings.ReplaceAll(suffix, " ", "_")
		tfName := fmt.Sprintf("customer_%s", suffix)

		f := hclwrite.NewEmptyFile()
		rootBody := f.Body()
		moduleBlock := rootBody.AppendNewBlock("module", []string{tfName})
		moduleBody := moduleBlock.Body()
		moduleBody.SetAttributeValue("source", cty.StringVal(source))
		moduleBody.SetAttributeValue("guid", cty.StringVal(guid))
		moduleBody.SetAttributeValue("name", cty.StringVal(name))
		moduleBody.SetAttributeValue("platform_environment", cty.StringVal(platformEnvironment))

		rootBody.AppendNewline()

		body := rootBody.AppendNewBlock("output", []string{tfName})
		body.Body().SetAttributeValue("value", cty.StringVal("module.customer_test_marka"))

		fmt.Printf("%s", f.Bytes())
	},
}

func init() {
	stubCustomerCMD.Flags().String("name", "", "Name of customer (readable)")
	stubCustomerCMD.Flags().String("platform-environment", "prod", "Platform environment (dev or prod), not linked to application environment")
	stubCustomerCMD.Flags().String("source", "./modules/dolittle-customer-with-resources", "Override with specific source of the module")
}
