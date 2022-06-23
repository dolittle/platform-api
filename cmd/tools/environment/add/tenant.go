package add

import (
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/pkg/namesgenerator"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var tenantCMD = &cobra.Command{
	Use:   "tenant",
	Short: "Adds a new tenant to the given application's environment",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logger := logrus.StandardLogger()
		logContext := logger.WithFields(logrus.Fields{
			"command": "environment add tenant",
		})

		applicationID, _ := cmd.Flags().GetString("application-id")
		environment, _ := cmd.Flags().GetString("environment")
		if applicationID == "" || environment == "" {
			logContext.Fatal("you have to specify the applicationID and environment")
		}

		tenantID, _ := cmd.Flags().GetString("tenant-id")
		if tenantID == "" {
			tenantID = uuid.New().String()
		}

		subdomain, _ := cmd.Flags().GetString("subdomain")
		if subdomain == "" {
			subdomain = namesgenerator.GetRandomName(0)
			subdomain = strings.ReplaceAll(subdomain, "_", "-")
		}
		host := fmt.Sprintf("%s.dolittle.cloud", subdomain)

		logContext = logContext.WithFields(logrus.Fields{
			"application_id": applicationID,
			"environment":    environment,
			"tenant_id":      tenantID,
			"subdomain":      subdomain,
			"host":           host,
		})

		k8sClient, _ := platformK8s.InitKubernetesClient()
	},
}

func init() {
	tenantCMD.Flags().String("application-id", "", "The application's ID")
	tenantCMD.Flags().String("environment", "", "The environment to add the tenant to")
	tenantCMD.Flags().String("tenant-id", "", "The tenant ID, defaults to a random one")
	tenantCMD.Flags().String("subdomain", "", "The subdomain for the tenant, defaults to a random one")
}
