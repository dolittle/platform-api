package automate

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/dolittle/platform-api/pkg/platform/automate"
)

var pullDolittleConfigCMD = &cobra.Command{
	Use:   "pull-dolittle-config",
	Short: "Pulls all dolittle configmaps from the cluster and writes them to their respective microservice inside the specified repo",
	Long: `
	go run main.go tools automate pull-dolittle-config <repo-root>
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
		ctx := context.TODO()
		kubeconfig := viper.GetString("tools.server.kubeConfig")

		if kubeconfig == "incluster" {
			kubeconfig = ""
		}

		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}

		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		namespaces := automate.GetNamespaces(ctx, client)
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

			configMaps, err := automate.GetDolittleConfigMaps(ctx, client, namespace.GetName())
			if err != nil {
				logContext.Fatal("Failed to get configmaps")
			}

			logContext.WithFields(logrus.Fields{
				"totalConfigMaps": len(configMaps),
			}).Info("Found dolittle configmaps")

			if dryRun {
				continue
			}

			err = automate.WriteConfigMapsToDirectory(args[0], configMaps)
			if err != nil {
				logContext.WithFields(logrus.Fields{
					"error": err,
				}).Fatal("Failed to write to disk")
			}
		}
	},
}

func init() {
	pullDolittleConfigCMD.PersistentFlags().Bool("dry-run", false, "Will not write to disk")
}
