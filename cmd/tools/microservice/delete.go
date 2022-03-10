package microservice

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deleteCMD = &cobra.Command{
	Use:   "delete",
	Short: "Delete microservices",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)
		logContext := logrus.StandardLogger()

		applicationID := viper.GetString("tools.microservice.application.id")
		environment := viper.GetString("tools.microservice.environment")
		microserviceID := viper.GetString("tools.microservice.id")
		logContext.Info(applicationID, environment, microserviceID)
	},
}
