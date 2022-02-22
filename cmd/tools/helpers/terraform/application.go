package terraform

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/spf13/cobra"
	"github.com/zclconf/go-cty/cty"
)

var applicationCMD = &cobra.Command{
	Use:   "application",
	Short: "Write terraform output for a customer application",
	Long: `
	Outputs a new Dolittle platform application in hcl to stdout.

	go run main.go tools helpers terraform application --application-name="Tree1" --application-id="fake-appliction-123" --customer="customer_test_marka"

	// Source/V3/Azure/{moduleName}.tf

	terraform init
	terraform apply -target="module.customer_test_marka_fake-application-123"
	`,
	Run: func(cmd *cobra.Command, args []string) {
		customer, _ := cmd.Flags().GetString("customer")
		applicationName, _ := cmd.Flags().GetString("application-name")
		applicationID, _ := cmd.Flags().GetString("application-id")

		if customer == "" {
			fmt.Println("A customer module is required")
			return
		}

		if applicationName == "" {
			fmt.Println("A customer application name is required")
			return
		}

		if applicationID == "" {
			applicationID = uuid.New().String()
			fmt.Printf("New customer application id generated: %s\n", applicationID)

		}

		source, _ := cmd.Flags().GetString("source")

		moduleName := fmt.Sprintf("%s_%s", customer, applicationID)

		customerApplication := tfCustomerApplication{
			Name:           applicationName,
			Source:         source,
			CustomerModule: customer,
			Guid:           applicationID,
			ModuleName:     moduleName,
		}

		f := hclwrite.NewEmptyFile()
		generateTerraformForCustomerApplication(f.Body(), customerApplication)

		fmt.Printf("%s", f.Bytes())
	},
}

type tfCustomerApplication struct {
	Source         string
	Guid           string
	Name           string
	ModuleName     string
	CustomerModule string
}

func generateTerraformForCustomerApplication(root *hclwrite.Body, customerApplication tfCustomerApplication) {
	/*
		module "customer_chris_taco" {
		  source     = "./modules/dolittle-application"
		  customer   = module.customer_chris
		  cluster_id = module.cluster_production_three.cluster_id
		  guid       = "11b6cf47-5d9f-438f-8116-0d9828654657"
		  name       = "Taco"
		}

		output "customer_chris_taco" {
		  value = module.customer_chris_taco
		}
	*/

	moduleBlock := root.AppendNewBlock("module", []string{customerApplication.ModuleName})
	moduleBody := moduleBlock.Body()
	moduleBody.SetAttributeValue("source", cty.StringVal(customerApplication.Source))
	moduleBody.SetAttributeTraversal("customer", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "module",
		},
		hcl.TraverseAttr{
			Name: customerApplication.CustomerModule,
		},
	})

	moduleBody.SetAttributeTraversal("cluster_id", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "module",
		},
		hcl.TraverseAttr{
			Name: "cluster_production_three.cluster_id",
		},
	})
	moduleBody.SetAttributeValue("guid", cty.StringVal(customerApplication.Guid))
	moduleBody.SetAttributeValue("name", cty.StringVal(customerApplication.Name))
	root.AppendNewline()

	output := root.AppendNewBlock("output", []string{customerApplication.ModuleName})
	output.Body().SetAttributeTraversal("value", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "module",
		},
		hcl.TraverseAttr{
			Name: customerApplication.ModuleName,
		},
	})

	output.Body().SetAttributeValue("sensitive", cty.True)
}

func init() {
	applicationCMD.Flags().String("application-name", "", "Name of application (readable)")
	applicationCMD.Flags().String("application-id", "", "Application ID to use")
	applicationCMD.Flags().String("customer", "", "Customer module name (customer_XXX)")
	applicationCMD.Flags().String("source", "./modules/dolittle-application", "Override with specific source of the module")
}
