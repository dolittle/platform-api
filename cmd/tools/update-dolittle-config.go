package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	configK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/tools/clientcmd"
)

var updateDolittleConfigCMD = &cobra.Command{
	Use:   "update-dolittle-config",
	Short: "Pulls all dolittle configmaps from the cluster and writes them to their respective microservice inside the specified repo",
	Long: `
	go run main.go tools update-dolittle-config <repo-root>
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logger := logrus.StandardLogger()

		//dryRun, _ := cmd.Flags().GetBool("dry-run")
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
					microservice := convertObjectMetaToMicroservice(configMap.GetObjectMeta())

					microservicePlatform := configK8s.NewMicroserviceConfigmapPlatformData(microservice)
					b, _ := json.MarshalIndent(microservicePlatform, "", "  ")
					platformJSON := string(b)
					configMap.Data["platform.json"] = platformJSON

					fmt.Println(configMap.Data)

					// client.CoreV1().ConfigMaps(namespace.GetName()).Update(ctx, &configMap, v1.UpdateOptions{})
				}
			}
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

		configMap.ManagedFields = nil

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

func init() {
	updateDolittleConfigCMD.PersistentFlags().Bool("dry-run", false, "Will not write to disk")
	updateDolittleConfigCMD.Flags().Bool("all", false, "To update all of the dolittle configmaps in the cluster")
}
