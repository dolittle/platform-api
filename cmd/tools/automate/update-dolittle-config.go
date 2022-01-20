package automate

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	configK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/tools/clientcmd"
)

var updateDolittleConfigCMD = &cobra.Command{
	Use:   "update-dolittle-config",
	Short: "Update xxx-config-dolittle",
	Long: `
	Update one or all dolittle configmaps, used by building blocks that have the runtime in use.

		go run main.go tools automate update-dolittle-config

		Via Stdin

		go run main.go tools automate get-microservices-metadata > ms.json 

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

		//k8sRepo := platform.NewK8sRepo(client, config)

		doAll, _ := cmd.Flags().GetBool("all")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		viaStdin, _ := cmd.Flags().GetBool("stdin")

		logContext := logger.WithFields(logrus.Fields{
			"dry_run": dryRun,
		})

		if doAll {
			namespaces := getNamespaces(ctx, client)
			for _, namespace := range namespaces {
				if !isApplicationNamespace(namespace) {
					continue
				}

				customer := namespace.Labels["tenant"]
				application := namespace.Labels["application"]
				logContext = logContext.WithFields(logrus.Fields{
					"customer":    customer,
					"application": application,
				})

				configMaps, err := getDolittleConfigMaps(ctx, client, namespace.GetName())

				if err != nil {
					logContext.Fatal("Failed to get configmaps")
				}

				logContext.WithFields(logrus.Fields{
					"totalConfigMaps": len(configMaps),
				}).Info("Found dolittle configmaps to update")

				for _, configMap := range configMaps {
					err := updateConfigMap(ctx, client, logContext, configMap, dryRun)
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

				microserviceMetadata, err := getOneConfigViaStdin(config)
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

	configMap, err := getOneDolittleConfigMap(ctx, client, applicationID, environment, microserviceID)
	if err != nil {
		logContext.Fatal("Failed to get configmap")
	}

	err = updateConfigMap(ctx, client, logContext, *configMap, dryRun)
	if err != nil {
		logContext.Fatal("Failed to update configmap")
	}
}

func getOneConfigViaStdin(input string) (MicroserviceMetadataShortInfo, error) {
	var data MicroserviceMetadataShortInfo
	err := json.Unmarshal([]byte(input), &data)
	return data, err
}

func getOneConfigViaParameters(cmd *cobra.Command) (applicationID string, environment string, microserviceID string) {
	applicationID, _ = cmd.Flags().GetString("application-id")
	environment, _ = cmd.Flags().GetString("environment")
	microserviceID, _ = cmd.Flags().GetString("microservice-id")
	return applicationID, environment, microserviceID
}

func updateConfigMap(ctx context.Context, client kubernetes.Interface, logContext logrus.FieldLogger, configMap corev1.ConfigMap, dryRun bool) error {
	microservice := convertObjectMetaToMicroservice(configMap.GetObjectMeta())
	platform := configK8s.NewMicroserviceConfigmapPlatformData(microservice)
	b, _ := json.MarshalIndent(platform, "", "  ")
	platformJSON := string(b)

	if configMap.Data == nil {
		// TODO this is a sign it might not be using a runtime, maybe we skip
		configMap.Data = make(map[string]string)
	}

	configMap.Data["platform.json"] = platformJSON

	namespace := configMap.Namespace

	logContext.WithFields(logrus.Fields{
		"microservice_id": microservice.ID,
		"application_id":  microservice.Application.ID,
		"environment":     microservice.Environment,
		"namespace":       microservice.Environment,
	})

	if dryRun {
		b := dumpConfigMap(&configMap)

		logContext = logContext.WithField("data", string(b))
		logContext.Info("Would write")
		return nil
	}
	fmt.Println("MISTAKE")
	return nil
	_, err := client.CoreV1().ConfigMaps(namespace).Update(ctx, &configMap, v1.UpdateOptions{})
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("updating configmap")
		return errors.New("update.failed")
	}
	return nil
}

func init() {
	updateDolittleConfigCMD.PersistentFlags().Bool("dry-run", false, "Will not write to disk")
	updateDolittleConfigCMD.Flags().Bool("all", false, "To update all of the dolittle configmaps in the cluster")
	updateDolittleConfigCMD.Flags().Bool("stdin", false, "Read from stdin")
	updateDolittleConfigCMD.Flags().String("application-id", "", "Application ID")
	updateDolittleConfigCMD.Flags().String("microservice-id", "", "Microservice ID")
	updateDolittleConfigCMD.Flags().String("environment", "", "environment")
}
