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
	Long: `

Add an event horizon subscription within an application and environment
go run main.go tools eventhorizon add 									\
	--customer-id 78a04863-376f-4279-9b52-d26d03279853 					\
	--application-id cfd9f00e-2308-0b4b-89dc-7b2280e19094 				\
	--environment Test 													\
	--microservice-id 76ba32bb-bbeb-45c5-925c-b8914cd1e6e4 				\
	--tenant-id 445f8ea8-1a6f-40d7-b2fc-796dba92dc44 					\
	--producer-microservice-id 6c70e222-1008-2e41-be43-bfe54f773ddd 	\
	--producer-tenant-id 445f8ea8-1a6f-40d7-b2fc-796dba92dc44 			\
	--public-stream 9a14541f-6077-40fc-bef5-89bce2c0790b 				\
	--partition whatever 												\
	--scope 443d87d1-ed23-4464-b321-de974a796a8c

Add an event horizon subscription across applications
go run main.go tools eventhorizon add 									\
	--customer-id 4cb310e8-8a8e-48a4-bb81-a8cddb484197					\
	--application-id e63b11dd-aaa6-2242-a0b9-230e4e06d43e 				\
	--environment Dev 													\
	--microservice-id e78f7323-19ff-445b-8dba-e0c17dadb569 				\
	--tenant-id d802ee55-644f-f142-af23-925391061fd6 					\
	--producer-application-id 35f0d011-f13f-8643-bea3-ef55ce47db2e 		\
	--producer-environment Dev										 	\
	--producer-microservice-id 418d5c62-5c6a-4b58-af51-93549d4635e5 	\
	--producer-tenant-id 43423763-e222-da40-bc77-bf275cc3875c 			\
	--public-stream 9a14541f-6077-40fc-bef5-89bce2c0790b 				\
	--partition whatever 												\
	--scope 443d87d1-ed23-4464-b321-de974a796a8c

	`,
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
