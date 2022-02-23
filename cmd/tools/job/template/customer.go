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

var customerCMD = &cobra.Command{
	Use:   "customer",
	Short: "Create a k8s Job to make customer",
	Long: `
	Outputs a k8s job to create a customers application


	go run main.go tools jobs create-customer \
	--customer-name="Test1"
	`,
	Run: func(cmd *cobra.Command, args []string) {
		customerID, _ := cmd.Flags().GetString("customer-id")
		if customerID == "" {
			customerID = uuid.New().String()
		}

		// TODO we shouldn't need this, but to re-use the labels we do
		// Get this from studio.json

		customerName, _ := cmd.Flags().GetString("customer-name")

		if customerName == "" {
			fmt.Println("--customer-name is required")
			return
		}

		customer := dolittleK8s.ShortInfo{
			ID:   customerID,
			Name: customerName,
		}

		createResourceConfig := jobK8s.CreateResourceConfigFromViper(viper.GetViper())
		resource := jobK8s.CreateCustomerResource(createResourceConfig, customer)

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
	customerCMD.Flags().String("customer-id", "", "Customer ID (optional, if not included, will create one)")
	customerCMD.Flags().String("customer-name", "", "Customer NAME")
}
