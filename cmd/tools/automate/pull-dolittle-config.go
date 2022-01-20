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

		scheme, serializer, err := initializeSchemeAndSerializer()
		if err != nil {
			panic(err.Error())
		}

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

func getMicroserviceDirectory(rootFolder string, configMap corev1.ConfigMap) string {
	labels := configMap.GetObjectMeta().GetLabels()
	customer := labels["tenant"]
	application := labels["application"]
	environment := labels["environment"]
	microservice := labels["microservice"]

	return filepath.Join(rootFolder,
		"Source",
		"V3",
		"Kubernetes",
		"Customers",
		customer,
		application,
		environment,
		microservice,
	)
}

func writeConfigMapsToDirectory(rootDirectory string, configMaps []corev1.ConfigMap, scheme *runtime.Scheme, serializer *k8sJson.Serializer) error {
	for _, configMap := range configMaps {
		// @joel let's discuss what to do with this
		// We remove these fields to make it cleaner and to make it a little less painful
		// to do multiple manual changes if we were debugging.
		configMap.ManagedFields = nil
		configMap.ResourceVersion = ""

		setConfigMapGVK(scheme, &configMap)

		microserviceDirectory := getMicroserviceDirectory(rootDirectory, configMap)
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

func setConfigMapGVK(schema *runtime.Scheme, configMap *corev1.ConfigMap) error {
	// get the GroupVersionKind of the configMap type from the schema
	gvks, _, err := schema.ObjectKinds(configMap)
	if err != nil {
		return err
	}
	// set the configMaps GroupVersionKind to match the one that the schema knows of
	configMap.GetObjectKind().SetGroupVersionKind(gvks[0])
	return nil
}

func initializeSchemeAndSerializer() (*runtime.Scheme, *k8sJson.Serializer, error) {
	// based of https://github.com/kubernetes/kubernetes/issues/3030#issuecomment-700099699
	// create the scheme and make it aware of the corev1 types
	scheme := runtime.NewScheme()
	err := corev1.AddToScheme(scheme)
	if err != nil {
		return scheme, nil, err
	}

	serializer := k8sJson.NewSerializerWithOptions(
		k8sJson.DefaultMetaFactory,
		scheme,
		scheme,
		k8sJson.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: true,
		},
	)
	return scheme, serializer, nil
}

func init() {
	pullDolittleConfigCMD.PersistentFlags().Bool("dry-run", false, "Will not write to disk")
}
