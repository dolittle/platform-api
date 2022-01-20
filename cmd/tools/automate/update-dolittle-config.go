package automate

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	configK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
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
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logger := logrus.StandardLogger()

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

		k8sRepo := platform.NewK8sRepo(client, config)

		scheme, serializer, err := initializeSchemeAndSerializer()
		if err != nil {
			panic(err.Error())
		}

		doAll, _ := cmd.Flags().GetBool("all")

		if doAll {
			namespaces := getNamespaces(ctx, client)
			for _, namespace := range namespaces {
				if !isApplicationNamespace(namespace) {
					continue
				}

				customer := namespace.Labels["tenant"]
				application := namespace.Labels["application"]
				logContext := logger.WithFields(logrus.Fields{
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
		return

		// Single lookup
		// TODO
		applicationID := "TODO"
		name := "TODO"
		configMap, err := k8sRepo.GetConfigMap(applicationID, name)
		if err != nil {
			panic(err.Error())
		}

		// TODO
		microservice := configK8s.Microservice{}
		platform := configK8s.NewMicroserviceConfigmapPlatformData(microservice)
		b, _ := json.MarshalIndent(platform, "", "  ")
		platformJSON := string(b)
		configMap.Data["platform.json"] = platformJSON

		// TODO not sure if we need to worry about this if we go straight back to k8s.
		//configMap.ManagedFields = nil
		//configMap.ResourceVersion = ""

		setConfigMapGVK(scheme, configMap)

		// TODO writes to stdout
		// TODO add flag to update the cluster
		// TODO add flag to write to file (for git)
		err = serializer.Encode(configMap, os.Stdout)
		if err != nil {
			panic(err.Error())
		}
	},
}

func updateConfigMap(ctx context.Context, client kubernetes.Interface, logContext logrus.FieldLogger, configMap corev1.ConfigMap, dryRun bool) error {
	microservice := convertObjectMetaToMicroservice(configMap.GetObjectMeta())
	platform := configK8s.NewMicroserviceConfigmapPlatformData(microservice)
	b, _ := json.MarshalIndent(platform, "", "  ")
	platformJSON := string(b)
	configMap.Data["platform.json"] = platformJSON

	namespace := configMap.Namespace

	logContext.WithFields(logrus.Fields{
		"microservice_id": microservice.ID,
		"application_id":  microservice.Application.ID,
		"environment":     microservice.Environment,
		"namespace":       microservice.Environment,
	})

	if dryRun {
		logContext.Info("Would write")
		return nil
	}

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
}
