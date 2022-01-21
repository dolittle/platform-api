package automate

import (
	"context"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/dolittle/platform-api/pkg/platform/automate"
	"k8s.io/apimachinery/pkg/runtime"
	k8sJson "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

var pullDolittleConfigCMD = &cobra.Command{
	Use:   "pull-dolittle-config",
	Short: "Pulls all dolittle configmaps from the cluster and writes them to their respective microservice inside the specified repo",
	Long: `
	go run main.go tools pull-dolittle-config <repo-root>
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

		scheme, serializer, err := automate.InitializeSchemeAndSerializerForConfigMap()
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

			err = writeConfigMapsToDirectory(args[0], configMaps, scheme, serializer)
			if err != nil {
				logContext.WithFields(logrus.Fields{
					"error": err,
				}).Fatal("Failed to write to disk")
			}
		}
	},
}

func writeConfigMapsToDirectory(rootDirectory string, configMaps []corev1.ConfigMap, scheme *runtime.Scheme, serializer *k8sJson.Serializer) error {
	for _, configMap := range configMaps {
		// We remove these fields to make it cleaner and to make it a little less painful
		// to do multiple manual changes if we were debugging.
		configMap.ManagedFields = nil
		configMap.ResourceVersion = ""

		automate.SetConfigMapGVK(scheme, &configMap)

		microserviceDirectory := automate.GetMicroserviceDirectory(rootDirectory, configMap)
		err := os.MkdirAll(microserviceDirectory, 0755)
		if err != nil {
			return err
		}

		file, err := os.Create(filepath.Join(microserviceDirectory, "microservice-configmap-dolittle.yml"))
		if err != nil {
			return err
		}

		defer file.Close()
		err = serializer.Encode(&configMap, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	pullDolittleConfigCMD.PersistentFlags().Bool("dry-run", false, "Will not write to disk")
}
