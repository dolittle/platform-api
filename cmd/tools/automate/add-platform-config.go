package automate

import (
	"bufio"
	"context"
	"encoding/json"
	"os"

	configK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/dolittle/platform-api/pkg/platform/automate"
	"k8s.io/client-go/tools/clientcmd"
)

var addPlatformConfigCMD = &cobra.Command{
	Use:   "add-platform-config",
	Short: "Adds platform.json to Dolittle microservices",
	Long: `
Add platform.json to one or all dolittle configmaps & Runtime containers volumeMounts.

# Update all
	go run main.go tools automate add-platform-config --all

# Update one via parameters

	go run main.go tools automate add-platform-config \
	--application-id="11b6cf47-5d9f-438f-8116-0d9828654657" \
	--environment="Dev" \
	--microservice-id="ec6a1a81-ed83-bb42-b82b-5e8bedc3cbc6" \
	--dry-run=true

# Update one or many via Stdin

	# Get metadata

	go run main.go tools automate get-microservices-metadata > ms.json

	# Filter and run a dry run

	cat ms.json| jq -c '.[]' | grep 'Nor-Sea' | grep 'Test' | \
	go run main.go tools automate add-platform-config --dry-run --stdin | \
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
			"dry_run":   dryRun,
			"do_all":    doAll,
			"via_stdin": viaStdin,
		})

		if doAll {
			microservices, err := automate.GetAllCustomerMicroservices(ctx, client)
			if err != nil {
				logContext.Fatal(err.Error())
			}

			for _, microservice := range microservices {
				logContext = logContext.WithFields(logrus.Fields{
					"customer":    microservice.Tenant.Name,
					"customer_id": microservice.Tenant.ID,
				})
				addPlatformDataToMicroservice(ctx, client, logContext, microservice.Application.ID, microservice.Environment, microservice.ID, dryRun)
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
				metadataJSON := scanner.Text()

				microserviceMetadata, err := automate.ParseMicroserviceMetadata(metadataJSON)
				logContext = logContext.WithFields(logrus.Fields{
					"customer":    microserviceMetadata.CustomerName,
					"customer_id": microserviceMetadata.CustomerID,
				})

				if err != nil {
					logContext.Fatal("Data via stdin is not valid")
				}

				applicationID = microserviceMetadata.ApplicationID
				environment = microserviceMetadata.Environment
				microserviceID = microserviceMetadata.MicroserviceID
				addPlatformDataToMicroservice(ctx, client, logContext, applicationID, environment, microserviceID, dryRun)
			}

			if scanner.Err() != nil {
				logContext.Error(scanner.Err())
				return
			}
			return
		}

		applicationID, environment, microserviceID = getMetadataViaFlags(cmd)
		addPlatformDataToMicroservice(ctx, client, logContext, applicationID, environment, microserviceID, dryRun)
	},
}

func addPlatformDataToMicroservice(ctx context.Context, client kubernetes.Interface, logContext logrus.FieldLogger, applicationID string, environment string, microserviceID string, dryRun bool) {
	logContext = logContext.WithFields(logrus.Fields{
		"application_id":  applicationID,
		"environment":     environment,
		"microservice_id": microserviceID,
	})

	configMap, err := automate.GetDolittleConfigMap(ctx, client, applicationID, environment, microserviceID)
	if err != nil {
		// TODO this might be to strict, perhaps we have a flag to let them skip?
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to get configmap")
	}

	microservice := automate.ConvertObjectMetaToMicroservice(configMap.GetObjectMeta())
	platform := configK8s.NewMicroserviceConfigMapPlatformData(microservice)
	platformJSON, _ := json.MarshalIndent(platform, "", "  ")

	err = automate.AddDataToConfigMap(ctx, client, logContext, "platform.json", platformJSON, *configMap, dryRun)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to update configmap")
	}

	containerIndex, deployment, err := automate.GetRuntimeDeployment(ctx, client, applicationID, environment, microserviceID)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to get runtime deployment")
	}

	platformMount := corev1.VolumeMount{
		Name:      "dolittle-config",
		MountPath: "/app/.dolittle/platform.json",
		SubPath:   "platform.json",
	}
	err = automate.AddVolumeMountToContainer(ctx, client, logContext, platformMount, containerIndex, deployment, dryRun)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to update runtime deployment")
	}
}

func getMetadataViaFlags(cmd *cobra.Command) (applicationID string, environment string, microserviceID string) {
	applicationID, _ = cmd.Flags().GetString("application-id")
	environment, _ = cmd.Flags().GetString("environment")
	microserviceID, _ = cmd.Flags().GetString("microservice-id")
	return applicationID, environment, microserviceID
}

func init() {
	addPlatformConfigCMD.PersistentFlags().Bool("dry-run", false, "Will not write to disk")
	addPlatformConfigCMD.Flags().Bool("all", false, "To update all of the dolittle configmaps in the cluster")
	addPlatformConfigCMD.Flags().Bool("stdin", false, "Read from stdin")
	addPlatformConfigCMD.Flags().String("application-id", "", "Application ID")
	addPlatformConfigCMD.Flags().String("microservice-id", "", "Microservice ID")
	addPlatformConfigCMD.Flags().String("environment", "", "environment")
}
