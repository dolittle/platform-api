package microservice

import (
	"context"
	"os"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	gitStorage "github.com/dolittle-entropy/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/thoas/go-funk"
	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var buildApplicationsCMD = &cobra.Command{
	Use:   "build-application-info",
	Short: "Write application info into the git repo",
	Long: `
	It will attempt to update git with data from the cluster and skip those that have been setup.

	GIT_REPO_SSH_KEY="/Users/freshteapot/dolittle/.ssh/test-deploy" \
	GIT_REPO_BRANCH=auto-dev \
	GIT_REPO_URL="git@github.com:freshteapot/test-deploy-key.git" \
	go run main.go microservice build-application-info --kube-config="/Users/freshteapot/.kube/config"
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logContext := logrus.StandardLogger()
		gitRepoConfig := initGit(logContext)

		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		ctx := context.TODO()
		kubeconfig, _ := cmd.Flags().GetString("kube-config")
		// TODO hoist localhost into viper
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}

		// create the clientset
		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		applications := extractApplications(ctx, client)
		SaveApplications(gitRepo, applications)
	},
}

func SaveApplications(repo storage.Repo, applications []platform.HttpResponseApplication) error {
	for _, application := range applications {
		studioConfig, _ := repo.GetStudioConfig(application.TenantID)
		if !studioConfig.BuildOverwrite {
			continue
		}
		err := repo.SaveApplication(application)
		if err != nil {
			panic(err.Error())
		}

		err = repo.SaveStudioConfig(application.TenantID, studioConfig)
		if err != nil {
			panic(err.Error())
		}
	}
	return nil
}

func extractApplications(ctx context.Context, client *kubernetes.Clientset) []platform.HttpResponseApplication {
	applications := make([]platform.HttpResponseApplication, 0)
	namespaces, err := client.CoreV1().Namespaces().List(ctx, metaV1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, ns := range namespaces.Items {
		// Unique the environments
		applications = append(applications, createApplicationFromK8s(ctx, client, ns))
	}

	return applications
}

func createApplicationFromK8s(ctx context.Context, client *kubernetes.Clientset, namespace v1.Namespace) platform.HttpResponseApplication {
	application := createBasicApplicationStructureFromMetadata(client, namespace)
	addEnvironmentsFromK8sInto(&application, ctx, client, namespace, application.TenantID, application.ID)
	return application
}

func createBasicApplicationStructureFromMetadata(client *kubernetes.Clientset, namespace v1.Namespace) platform.HttpResponseApplication {
	annotationsMap := namespace.GetObjectMeta().GetAnnotations()
	labelMap := namespace.GetObjectMeta().GetLabels()

	applicationID := annotationsMap["dolittle.io/application-id"]
	tenantID := annotationsMap["dolittle.io/tenant-id"]

	return platform.HttpResponseApplication{
		TenantID:     tenantID,
		ID:           applicationID,
		Name:         labelMap["application"],
		Environments: make([]platform.HttpInputEnvironment, 0),
	}
}

func addEnvironmentsFromK8sInto(application *platform.HttpResponseApplication, ctx context.Context, client *kubernetes.Clientset, namespace v1.Namespace, tenantID, applicationID string) {
	namespaceName := namespace.GetObjectMeta().GetName()
	for _, deployment := range getDeployments(ctx, client, namespaceName) {
		if !deploymentHasReqiredMetadata(deployment) {
			continue
		}

		environment := createEnvironmentFromK8s(deployment, tenantID, applicationID)

		if environmentAlreadyInApplication(application.Environments, environment.Name) {
			continue
		}

		application.Environments = append(application.Environments, environment)
	}
}

func getDeployments(ctx context.Context, client *kubernetes.Clientset, namespace string) []appsV1.Deployment {
	deploymentList, err := client.AppsV1().Deployments(namespace).List(ctx, metaV1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	return deploymentList.Items
}

func createEnvironmentFromK8s(deployment appsV1.Deployment, tenantID, applicationID string) platform.HttpInputEnvironment {
	return platform.HttpInputEnvironment{
		Name:          deployment.ObjectMeta.Labels["environment"],
		TenantID:      tenantID,
		ApplicationID: applicationID,
	}
}
func deploymentHasReqiredMetadata(deployment appsV1.Deployment) bool {
	_, ok := deployment.ObjectMeta.Annotations["dolittle.io/tenant-id"]
	if !ok {
		return false
	}

	_, ok = deployment.ObjectMeta.Annotations["dolittle.io/application-id"]
	if !ok {
		return false
	}

	_, ok = deployment.ObjectMeta.Labels["application"]
	if !ok {
		return false
	}
	return true
}
func environmentAlreadyInApplication(environments []platform.HttpInputEnvironment, environmentName string) bool {
	index := funk.IndexOf(environments, func(item platform.HttpInputEnvironment) bool {
		return item.Name == environmentName
	})

	if index != -1 {
		return false
	}
	return true
}
func init() {
	RootCmd.AddCommand(buildApplicationsCMD)
	buildApplicationsCMD.Flags().String("kube-config", "", "FullPath to kubeconfig")
}
