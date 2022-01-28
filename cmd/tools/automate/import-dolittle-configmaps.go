package automate

import (
	"bufio"
	"context"
	"os"

	"github.com/dolittle/platform-api/pkg/platform/automate"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var importDolittleConfigMapsCMD = &cobra.Command{
	Use:   "import-dolittle-configmaps",
	Short: "Creates dolittle configmaps and namespaces from the given JSON configmaps",
	Long: `
Creates all of the given dolittle configmaps and their namespaces from the given JSON so that it's easy to populate a development cluster.

The JSON has to be created in the following format:
	kubectl get cm -A -l microservice -o json | jq -c '.items[] | select(.metadata.name | test(".+-dolittle$"))' > configmaps.ndjson

NOTE! MAKE SURE TO REMEMBER TO CHANGE YOUR CONTEXT:
	kubectl config use-context k3d-dolittle-dev

Then you can feed it to the command:
	cat configmaps.ndjson | go run main.go tools automate import-dolittle-configmaps
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logger := logrus.StandardLogger()

		dryRun, _ := cmd.Flags().GetBool("dry-run")

		k8sClient, _ := platformK8s.InitKubernetesClient()

		scheme, serializer, err := automate.InitializeSchemeAndSerializer()
		if err != nil {
			panic(err.Error())
		}

		namespaces := make(map[string]corev1.Namespace)
		var configMaps []*corev1.ConfigMap

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			jsonCM := scanner.Bytes()

			runtimeConfigMap := &corev1.ConfigMap{}
			gvks, _, err := scheme.ObjectKinds(runtimeConfigMap)
			if err != nil {
				logger.Fatal(err)
			}
			_, _, err = serializer.Decode(jsonCM, &gvks[0], runtimeConfigMap)
			if err != nil {
				logger.Fatal(err)
			}

			// ResourceVersion should not be se on object to be created
			runtimeConfigMap.ResourceVersion = ""
			runtimeConfigMap.ManagedFields = nil

			namespaceName := runtimeConfigMap.GetObjectMeta().GetNamespace()
			configMaps = append(configMaps, runtimeConfigMap)

			if _, ok := namespaces[namespaceName]; !ok {
				namespaces[namespaceName] = corev1.Namespace{
					TypeMeta: v1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Namespace",
					},
					ObjectMeta: v1.ObjectMeta{
						Name: namespaceName,
						Labels: map[string]string{
							"tenant":      runtimeConfigMap.Labels["tenant"],
							"application": runtimeConfigMap.Labels["application"],
						},
						Annotations: map[string]string{
							"dolittle.io/tenant-id":      runtimeConfigMap.Annotations["dolittle.io/tenant-id"],
							"dolittle.io/application-id": runtimeConfigMap.Annotations["dolittle.io/application-id"],
						},
					},
				}
			}
		}

		createNamespaces(k8sClient, namespaces, dryRun, logger)
		createConfigMaps(k8sClient, configMaps, dryRun, logger)
	},
}

func createNamespaces(client kubernetes.Interface, namespaces map[string]corev1.Namespace, dryRun bool, logger logrus.FieldLogger) {
	ctx := context.TODO()
	for name, namespace := range namespaces {
		logContext := logger.WithFields(logrus.Fields{
			"function":    "createNamespaces",
			"namespace":   name,
			"customer":    namespace.Labels["tenant"],
			"application": namespace.Labels["application"],
		})
		if dryRun {
			logContext.Infof("Would've created namespace %s", name)
			continue
		}
		_, err := client.CoreV1().Namespaces().Create(ctx, &namespace, v1.CreateOptions{})
		if err != nil {
			if !k8serrors.IsAlreadyExists(err) {
				logContext.Fatal(err)
			}
			logContext.Infof("Namespace %s already exists", name)
		} else {
			logContext.Infof("Created namespace %s", name)
		}
	}
}

func createConfigMaps(client kubernetes.Interface, configMaps []*corev1.ConfigMap, dryRun bool, logger logrus.FieldLogger) {

	ctx := context.TODO()
	for _, configMap := range configMaps {
		logContext := logger.WithFields(logrus.Fields{
			"function":     "createConfigMaps",
			"namespace":    configMap.Namespace,
			"customer":     configMap.Labels["tenant"],
			"application":  configMap.Labels["application"],
			"microservice": configMap.Labels["microservice"],
			"environment":  configMap.Labels["environment"],
			"configMap":    configMap.Name,
		})
		if dryRun {
			logContext.Infof("Would've created ConfigMap %s", configMap.Name)
			continue
		}
		_, err := client.CoreV1().ConfigMaps(configMap.Namespace).Create(ctx, configMap, v1.CreateOptions{})
		if err != nil {
			if !k8serrors.IsAlreadyExists(err) {
				panic(err.Error())
			}
			logContext.Infof("ConfigMap %s already exists", configMap.Name)
		} else {
			logContext.Infof("Created ConfigMap %s", configMap.Name)
		}
	}
}

func init() {
	importDolittleConfigMapsCMD.PersistentFlags().Bool("dry-run", false, "Will not update the cluster")
}
