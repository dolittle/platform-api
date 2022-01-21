package automate

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"

	"github.com/dolittle/platform-api/pkg/platform/automate"
	"k8s.io/client-go/tools/clientcmd"
)

var updateDolittleConfigCMD = &cobra.Command{
	Use:   "update-dolittle-config",
	Short: "Update xxx-config-dolittle",
	Long: `
Update one or all dolittle configmaps, used by building blocks that have the runtime in use.

# Update all
	go run main.go tools automate update-dolittle-config --all

# Update one via parameters

	go run main.go tools automate update-dolittle-config \
	--application-id="11b6cf47-5d9f-438f-8116-0d9828654657" \
	--environment="Dev" \
	--microservice-id="ec6a1a81-ed83-bb42-b82b-5e8bedc3cbc6" \
	--dry-run=true

# Update one or many via Stdin

	# Get metadata

	go run main.go tools automate get-microservices-metadata > ms.json

	# Filter and run a dry run

	cat ms.json| jq -c '.[]' | grep 'Nor-Sea' | grep 'Test' | \
	go run main.go tools automate update-dolittle-config --dry-run --stdin | \
	jq -r '.data' | \
	yq e -
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logger := logrus.StandardLogger()

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

		doAll, _ := cmd.Flags().GetBool("all")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		viaStdin, _ := cmd.Flags().GetBool("stdin")

		logContext := logger.WithFields(logrus.Fields{
			"dry_run": dryRun,
		})

		if doAll {
			namespaces := automate.GetNamespaces(ctx, client)
			for _, namespace := range namespaces {
				if !automate.IsApplicationNamespace(namespace) {
					continue
				}

				customer := namespace.Labels["tenant"]
				application := namespace.Labels["application"]
				logContext = logContext.WithFields(logrus.Fields{
					"customer":    customer,
					"application": application,
				})

				configMaps, err := automate.GetDolittleConfigMaps(ctx, client, namespace.GetName())

				if err != nil {
					// TODO this might be to strict, perhaps we have a flag to let them skip?
					logContext.Fatal("Failed to get configmaps")
				}

				logContext.WithFields(logrus.Fields{
					"totalConfigMaps": len(configMaps),
				}).Info("Found dolittle configmaps to update")

				for _, configMap := range configMaps {
					err := automate.UpdateConfigMap(ctx, client, logContext, configMap, dryRun)
					if err != nil {
						logContext.Fatal("Failed to update configmap")
					}
				}
			}
			return
		}

		var (
			applicationID  string
			environment    string
			microserviceID string
		)

		if viaStdin {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				config := scanner.Text()

				microserviceMetadata, err := automate.GetOneConfigViaStdin(config)
				if err != nil {
					logContext.Fatal("Data via stdin is not valid")
				}

				applicationID = microserviceMetadata.ApplicationID
				environment = microserviceMetadata.Environment
				microserviceID = microserviceMetadata.MicroserviceID
				processOne(ctx, client, logContext, dryRun, applicationID, environment, microserviceID)
			}

			if scanner.Err() != nil {
				// Handle error.
				fmt.Println(scanner.Err())
				return
			}
			return
		}

		applicationID, environment, microserviceID = getOneConfigViaParameters(cmd)
		processOne(ctx, client, logContext, dryRun, applicationID, environment, microserviceID)

	},
}

func processOne(ctx context.Context, client kubernetes.Interface, logContext logrus.FieldLogger, dryRun bool, applicationID string, environment string, microserviceID string) {
	logContext = logContext.WithFields(logrus.Fields{
		"application_id":   applicationID,
		"environment":      environment,
		"microservivce_id": microserviceID,
	})

	configMap, err := automate.GetOneDolittleConfigMap(ctx, client, applicationID, environment, microserviceID)
	if err != nil {
		// TODO this might be to strict, perhaps we have a flag to let them skip?
		logContext.Fatal("Failed to get configmap")
	}

	err = automate.UpdateConfigMap(ctx, client, logContext, *configMap, dryRun)
	if err != nil {
		logContext.Fatal("Failed to update configmap")
	}
}

func getOneConfigViaParameters(cmd *cobra.Command) (applicationID string, environment string, microserviceID string) {
	applicationID, _ = cmd.Flags().GetString("application-id")
	environment, _ = cmd.Flags().GetString("environment")
	microserviceID, _ = cmd.Flags().GetString("microservice-id")
	return applicationID, environment, microserviceID
}

func init() {
	updateDolittleConfigCMD.PersistentFlags().Bool("dry-run", false, "Will not write to disk")
	updateDolittleConfigCMD.Flags().Bool("all", false, "To update all of the dolittle configmaps in the cluster")
	updateDolittleConfigCMD.Flags().Bool("stdin", false, "Read from stdin")
	updateDolittleConfigCMD.Flags().String("application-id", "", "Application ID")
	updateDolittleConfigCMD.Flags().String("microservice-id", "", "Microservice ID")
	updateDolittleConfigCMD.Flags().String("environment", "", "environment")
}
