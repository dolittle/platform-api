package create

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dolittle/platform-api/pkg/aiven"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/m3connector"
	"github.com/dolittle/platform-api/pkg/platform/microservice/m3connector/k8s"
)

var environmentCMD = &cobra.Command{
	Use:   "environment",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logger := logrus.StandardLogger()
		logContext := logger.WithFields(logrus.Fields{
			"command": "environment",
		})

		apiToken := viper.GetString("tools.m3connector.aiven.apiToken")
		if apiToken == "" {
			logContext.Fatal("you have to provide an Aiven api token")
		}

		project := viper.GetString("tools.m3connector.aiven.project")
		service := viper.GetString("tools.m3connector.aiven.service")
		if project == "" || service == "" {
			logContext.Fatal("you have to specify both the project and the service")
		}

		customerID, _ := cmd.Flags().GetString("customer-id")
		applicationID, _ := cmd.Flags().GetString("application-id")
		environment, _ := cmd.Flags().GetString("environment")
		if customerID == "" || applicationID == "" || environment == "" {
			logContext.Fatal("you have to specify the customerID, applicationID and environment")
		}

		isProduction := viper.GetBool("tools.server.isProduction")

		logContext = logContext.WithFields(logrus.Fields{
			"customer_id":    customerID,
			"application_id": applicationID,
			"environment":    environment,
		})

		aiven, err := aiven.NewClient(apiToken, project, service, logContext)
		if err != nil {
			logContext.Fatal(err)
		}

		k8sClient, _ := platformK8s.InitKubernetesClient()

		k8sRepo := k8s.NewM3ConnectorRepo(k8sClient, isProduction, logContext.Logger)

		m3connector := m3connector.NewM3Connector(aiven, k8sRepo, logContext)
		err = m3connector.CreateEnvironment(customerID, applicationID, environment)
		if err != nil {
			logContext.Fatal(err)
		}
		logContext.Info("done")
	},
}

func init() {
	environmentCMD.Flags().String("customer-id", "", "The customers ID")
	environmentCMD.Flags().String("application-id", "", "The applications ID")
	environmentCMD.Flags().String("environment", "", "The environment")
}
