package automate

import (
	"bufio"
	"context"
	"encoding/json"
	"os"

	configK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/dolittle/platform-api/pkg/platform/automate"
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
		k8sClient, _ := platformK8s.InitKubernetesClient()
		k8sRepoV2 := k8s.NewRepo(k8sClient, logger.WithField("context", "k8s-repo-v2"))

		doAll, _ := cmd.Flags().GetBool("all")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		viaStdin, _ := cmd.Flags().GetBool("stdin")

		logContext := logger.WithFields(logrus.Fields{
			"dry_run":   dryRun,
			"do_all":    doAll,
			"via_stdin": viaStdin,
		})

		if doAll {
			microservices, err := automate.GetAllCustomerMicroservices(k8sRepoV2)
			if err != nil {
				logContext.Fatal(err.Error())
			}

			for _, microservice := range microservices {
				logContext = logContext.WithFields(logrus.Fields{
					"customer":     microservice.Tenant.Name,
					"customer_id":  microservice.Tenant.ID,
					"microservice": microservice.Name,
					"application":  microservice.Application.Name,
				})
				addPlatformDataToMicroservice(ctx, k8sClient, logContext, microservice.Application.ID, microservice.Environment, microservice.ID, dryRun)
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
					"customer":          microserviceMetadata.CustomerName,
					"customer_id":       microserviceMetadata.CustomerID,
					"microservice_name": microserviceMetadata.MicroserviceName,
					"application":       microserviceMetadata.ApplicationName,
				})

				if err != nil {
					logContext.Fatal("Data via stdin is not valid")
				}

				applicationID = microserviceMetadata.ApplicationID
				environment = microserviceMetadata.Environment
				microserviceID = microserviceMetadata.MicroserviceID
				addPlatformDataToMicroservice(ctx, k8sClient, logContext, applicationID, environment, microserviceID, dryRun)
			}

			if scanner.Err() != nil {
				logContext.Error(scanner.Err())
				return
			}
			return
		}

		applicationID, environment, microserviceID = getMetadataViaFlags(cmd)
		addPlatformDataToMicroservice(ctx, k8sClient, logContext, applicationID, environment, microserviceID, dryRun)
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
		if err != platform.ErrNotFound {
			logContext.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Failed to get dolittle configmap")
		}

		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Info("No configmap found")
		return
	}

	// here we can add the missing names if it wasn't already added, like when figuring out from CLI flags
	logContext = logContext.WithFields(logrus.Fields{
		"microservice": configMap.Labels["microservice"],
		"application":  configMap.Labels["application"],
	})

	microservice := automate.ConvertObjectMetaToMicroservice(configMap.GetObjectMeta())
	platformData := configK8s.NewMicroserviceConfigMapPlatformData(microservice)
	platformJSON, _ := json.MarshalIndent(platformData, "", "  ")

	err = automate.AddDataToConfigMap(ctx, client, logContext, "platform.json", platformJSON, *configMap, dryRun)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to update configmap")
	}

	deployment, err := automate.GetDeployment(ctx, client, applicationID, environment, microserviceID)
	if err != nil {
		if err != platform.ErrNotFound {
			logContext.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Failed to get deployment")
		}

		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Info("No deployment found")
		return
	}
	runtimeContainerIndex := automate.GetContainerIndex(deployment, "runtime")
	if runtimeContainerIndex == -1 {
		logContext.Error("deployment didn't have a runtime container")
		return
	}

	platformMount := corev1.VolumeMount{
		Name:      "dolittle-config",
		MountPath: "/app/.dolittle/platform.json",
		SubPath:   "platform.json",
	}
	err = automate.AddVolumeMountToContainer(ctx, client, logContext, platformMount, runtimeContainerIndex, deployment, dryRun)
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
