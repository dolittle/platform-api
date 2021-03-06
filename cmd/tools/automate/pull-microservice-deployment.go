package automate

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform/automate"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
)

var pullMicroserviceDeploymentCMD = &cobra.Command{
	Use:   "pull-microservice-deployment",
	Short: "Pulls all dolittle microservice deployments from the cluster and writes them to their respective microservice inside the specified repo",
	Long: `
	go run main.go tools automate pull-microservice-deployment <repo-root>
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logger := logrus.StandardLogger()

		if len(args) == 0 {
			logger.Error("Specify the directory to write to")
			return
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		k8sClient, _ := platformK8s.InitKubernetesClient()

		k8sRepoV2 := k8s.NewRepo(k8sClient, logger.WithField("context", "k8s-repo-v2"))

		namespaces, _ := k8sRepoV2.GetNamespacesWithApplication()
		for _, namespace := range namespaces {
			if !automate.IsApplicationNamespace(namespace) {
				continue
			}
			customer := namespace.Labels["tenant"]
			application := namespace.Labels["application"]
			logContext := logger.WithFields(logrus.Fields{
				"customer":    customer,
				"application": application,
			})

			deployments, err := k8sRepoV2.GetDeploymentsWithMicroservice(namespace.Name)
			if err != nil {
				logContext.Fatal("Failed to get deployments")
			}

			logContext.WithFields(logrus.Fields{
				"totalDeployments": len(deployments),
			}).Info("Found microservice deployments")

			if dryRun {
				continue
			}

			err = automate.WriteDeploymentsToDirectory(args[0], deployments)
			if err != nil {
				logContext.WithFields(logrus.Fields{
					"error": err,
				}).Fatal("Failed to write to disk")
			}
		}
	},
}

func init() {
	pullMicroserviceDeploymentCMD.PersistentFlags().Bool("dry-run", false, "Will not write to disk")
}
