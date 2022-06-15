package eventhorizon

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var addCMD = &cobra.Command{
	Use:   "add",
	Short: "add a new event horizon subscription",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)
		logger := logrus.StandardLogger()
		logContext := logger.WithFields(logrus.Fields{
			"command": "eventhorizon add",
		})

		customerID, _ := cmd.Flags().GetString("customer-id")
		applicationID, _ := cmd.Flags().GetString("application-id")
		environment, _ := cmd.Flags().GetString("environment")
		microserviceID, _ := cmd.Flags().GetString("microservice-id")
		tenantID, _ := cmd.Flags().GetString("tenant-id")
		producerMicroserviceID, _ := cmd.Flags().GetString("producer-microservice-id")
		producerTenantID, _ := cmd.Flags().GetString("producer-tenant-id")
		publicStream, _ := cmd.Flags().GetString("public-stream")
		partition, _ := cmd.Flags().GetString("partition")

		if customerID == "" ||
			applicationID == "" ||
			environment == "" ||
			microserviceID == "" ||
			tenantID == "" ||
			producerMicroserviceID == "" ||
			producerTenantID == "" ||
			publicStream == "" ||
			partition == "" {
			logContext.Fatal("you have to specify the required flags")
		}

		logContext = logContext.WithFields(logrus.Fields{
			"customer_id":              customerID,
			"application_id":           applicationID,
			"environment":              environment,
			"microservice_id":          microserviceID,
			"consumer_tenant_id":       tenantID,
			"producer_microservice_id": producerMicroserviceID,
			"producer_tenant_id":       producerTenantID,
			"public_stream":            publicStream,
			"partition":                partition,
		})

		// k8sClient, _ := platformK8s.InitKubernetesClient()
	},
}

func init() {
	addCMD.Flags().String("customer-id", "", "The customers ID. Both consumer and producer have to be owned by the same customer")
	addCMD.Flags().String("application-id", "", "The consumers applications ID")
	addCMD.Flags().String("microservice-id", "", "The consumers microservice ID")
	addCMD.Flags().String("environment", "", "The consumers environment")
	addCMD.Flags().String("tenant-id", "", "The consumers tenantID")

	addCMD.Flags().String("producer-microservice-id", "", "The producers microservice ID")
	addCMD.Flags().String("producer-tenant-id", "", "The producers tenant that gives consent to it's public stream")
	addCMD.Flags().String("public-stream", "", "The producers public stream to subscribe to")
	addCMD.Flags().String("partition", "", "The producers public streams partition to subscribe to")

	addCMD.Flags().String("producer-application-id", "", "(optional) The producers applications ID if different from the consumer")
	addCMD.Flags().String("producer-environment", "", "(optional) The produecers environment if different from the consumer")
}
