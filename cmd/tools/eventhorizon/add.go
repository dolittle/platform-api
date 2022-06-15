package eventhorizon

import (
	"os"

	"github.com/dolittle/platform-api/pkg/k8s"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	k8sSimple "github.com/dolittle/platform-api/pkg/platform/microservice/simple/k8s"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		scope, _ := cmd.Flags().GetString("scope")

		if customerID == "" ||
			applicationID == "" ||
			environment == "" ||
			microserviceID == "" ||
			tenantID == "" ||
			producerMicroserviceID == "" ||
			producerTenantID == "" ||
			publicStream == "" ||
			partition == "" ||
			scope == "" {
			logContext.Fatal("you have to specify the required flags")
		}

		producerApplicationID, _ := cmd.Flags().GetString("producer-application-id")
		producerEnvironment, _ := cmd.Flags().GetString("producer-environment")

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
			"scope":                    scope,
			"producer_application_id":  producerApplicationID,
			"producer_environment":     producerEnvironment,
		})

		k8sClient, config := platformK8s.InitKubernetesClient()
		k8sRepo := platformK8s.NewK8sRepo(k8sClient, config, logContext)
		k8sRepoV2 := k8s.NewRepo(k8sClient, logContext)

		isProduction := viper.GetBool("tools.server.isProduction")
		simpleRepo := k8sSimple.NewSimpleRepo(k8sClient, k8sRepo, k8sRepoV2, isProduction)

		if producerApplicationID == "" || producerEnvironment == "" {
			// do the simple case where procucer & consumer are in the same application and same environment
			err := simpleRepo.Subscribe(
				customerID,
				applicationID,
				environment,
				microserviceID,
				tenantID,
				producerMicroserviceID,
				producerTenantID,
				publicStream,
				partition,
				scope)
			if err != nil {
				logContext.Fatal(err)
			}
		} else {
			err := simpleRepo.SubscribeToAnotherApplication(
				customerID,
				applicationID,
				environment,
				microserviceID,
				tenantID,
				producerMicroserviceID,
				producerTenantID,
				publicStream,
				partition,
				scope,
				producerApplicationID,
				producerEnvironment)
			if err != nil {
				logContext.Fatal(err)
			}
		}
		logContext.Info("job done")
	},
}

func init() {
	addCMD.Flags().String("customer-id", "", "The customers ID. Both consumer and producer have to be owned by the same customer")
	addCMD.Flags().String("application-id", "", "The consumers applications ID")
	addCMD.Flags().String("microservice-id", "", "The consumers microservice ID")
	addCMD.Flags().String("environment", "", "The consumers environment")
	addCMD.Flags().String("tenant-id", "", "The consumers tenantID")
	addCMD.Flags().String("scope", "", "The consumers scope to put the events to")

	addCMD.Flags().String("producer-microservice-id", "", "The producers microservice ID")
	addCMD.Flags().String("producer-tenant-id", "", "The producers tenant that gives consent to it's public stream")
	addCMD.Flags().String("public-stream", "", "The producers public stream to subscribe to")
	addCMD.Flags().String("partition", "", "The producers public streams partition to subscribe to")

	addCMD.Flags().String("producer-application-id", "", "(optional) The producers applications ID if different from the consumer")
	addCMD.Flags().String("producer-environment", "", "(optional) The produecers environment if different from the consumer")
}
