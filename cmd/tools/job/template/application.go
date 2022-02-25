package template

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

var applicationCMD = &cobra.Command{
	Use:   "application",
	Short: "Create a k8s Job to make a customers application",
	Long: `
	Outputs a k8s job to create a customers application

	go run main.go tools job template application \
	--platform-environment="dev" \
	--application-name="Tree1" \
	--application-id="fake-appliction-123" \
	--customer-id=""
	`,
	Run: func(cmd *cobra.Command, args []string) {

		customerID, _ := cmd.Flags().GetString("customer-id")
		// TODO we shouldn't need this, but to re-use the labels we do
		// Get this from studio.json
		// If we use config file, then we can drop these
		platformEnvironment := viper.GetString("tools.server.platformEnvironment")
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

		createResourceConfig := jobK8s.CreateResourceConfigFromViper(viper.GetViper())
		// if we move to config file, we can drop these
		createResourceConfig.PlatformEnvironment = platformEnvironment
		createResourceConfig.IsProduction = isProduction

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

		serializer.Encode(resource, os.Stdout)
	},
}

func init() {
	applicationCMD.Flags().String("application-name", "", "Name of application (readable)")
	applicationCMD.Flags().String("application-id", "", "Application ID to use")
	applicationCMD.Flags().String("customer-id", "", "Customer ID")
	applicationCMD.Flags().Bool("is-production", false, "Signal this is in production mode")
}
