package jobs

import (
	"fmt"
	"os"

	k8sJson "k8s.io/apimachinery/pkg/runtime/serializer/json"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	jobK8s "github.com/dolittle/platform-api/pkg/platform/job/k8s"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8s.io/apimachinery/pkg/runtime"
)

var createCustomerApplicationCMD = &cobra.Command{
	Use:   "create-customer-application",
	Short: "Create job to make customer application",
	Long: `
	Outputs a new Dolittle platform applicaiton in hcl to stdout.

	go run main.go tools jobs create-customer-application \
	--platform-environment="dev" \
	--application-name="Tree1" \
	--application-id="fake-appliction-123" \
	--customer-id=""
	`,
	Run: func(cmd *cobra.Command, args []string) {
		platformOperationsImage := viper.GetString("tools.jobs.image.operations")
		customerID, _ := cmd.Flags().GetString("customer-id")
		// TODO we shouldn't need this, but to re-use the labels we do
		// Get this from studio.json
		platformEnvironment, _ := cmd.Flags().GetString("platform-environment")
		isProduction, _ := cmd.Flags().GetBool("is-production")
		applicationName, _ := cmd.Flags().GetString("application-name")
		applicationID, _ := cmd.Flags().GetString("application-id")

		if customerID == "" {
			fmt.Println("--customer-id is required")
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

		application := dolittleK8s.ShortInfo{
			ID:   applicationID,
			Name: applicationName,
		}

		createResourceConfig := jobK8s.CreateResourceConfigWithDefaults(platformOperationsImage, platformEnvironment, isProduction)
		resource := jobK8s.CreateApplicationResource(createResourceConfig, customerID, application)

		s := runtime.NewScheme()
		serializer := k8sJson.NewSerializerWithOptions(
			k8sJson.DefaultMetaFactory,
			s,
			s,
			k8sJson.SerializerOptions{
				Yaml:   true,
				Pretty: true,
				Strict: true,
			},
		)

		dryRun := true
		if dryRun {
			serializer.Encode(resource, os.Stdout)
			return
		}
	},
}

func init() {
	createCustomerApplicationCMD.Flags().String("application-name", "", "Name of application (readable)")
	createCustomerApplicationCMD.Flags().String("application-id", "", "Application ID to use")
	createCustomerApplicationCMD.Flags().String("customer-id", "", "Customer ID")
	createCustomerApplicationCMD.Flags().String("platform-environment", "dev", "Platform environment (dev or prod), not linked to application environment")
	createCustomerApplicationCMD.Flags().Bool("is-production", false, "Signal this is in production mode")
}
