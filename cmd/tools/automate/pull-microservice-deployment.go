package automate

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/dolittle/platform-api/pkg/platform/automate"
	"k8s.io/apimachinery/pkg/runtime"
	k8sJson "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

var pullMicroserviceDeploymentCMD = &cobra.Command{
	Use:   "pull-microservice-deployment",
	Short: "Pulls all dolittle microservice deployments from the cluster and writes them to their respective microservice inside the specified repo",
	Long: `
	go run main.go tools pull-microservice-deployment <repo-root>
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

		scheme, serializer, err := automate.InitializeSchemeAndSerializer()
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

			deployments, err := automate.GetDeployments(ctx, client, namespace.GetName())
			if err != nil {
				logContext.Fatal("Failed to get deployments")
			}

			logContext.WithFields(logrus.Fields{
				"totalDeployments": len(deployments),
			}).Info("Found microservice deployments")

			if dryRun {
				continue
			}

			err = writeDeploymentsToDirectory(args[0], deployments, scheme, serializer)
			if err != nil {
				logContext.WithFields(logrus.Fields{
					"error": err,
				}).Fatal("Failed to write to disk")
			}
		}
	},
}

func writeDeploymentsToDirectory(rootDirectory string, deployments []appsv1.Deployment, scheme *runtime.Scheme, serializer *k8sJson.Serializer) error {
	for _, deployment := range deployments {
		// We remove these fields to make it cleaner and to make it a little less painful
		// to do multiple manual changes if we were debugging.
		deployment.ManagedFields = nil
		deployment.ResourceVersion = ""
		deployment.Status = appsv1.DeploymentStatus{}
		delete(deployment.ObjectMeta.Annotations, "kubectl.kubernetes.io/last-applied-configuration")

		automate.SetRuntimeObjectGVK(scheme, &deployment)

		microserviceDirectory := automate.GetMicroserviceDirectory(rootDirectory, deployment.GetObjectMeta())
		automate.WriteResourceToFile(microserviceDirectory, "microservice-deployment.yml", &deployment, serializer)
	}

	return nil
}

func init() {
	pullMicroserviceDeploymentCMD.PersistentFlags().Bool("dry-run", false, "Will not write to disk")
}
