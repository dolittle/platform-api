package microservice

import (
	"context"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	gitStorage "github.com/dolittle-entropy/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/thoas/go-funk"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var buildApplicationsCMD = &cobra.Command{
	Use:   "build-application-info",
	Short: "Write application info into the git repo",
	Long: `
	It will attempt to update git with data from the cluster and skip those that have been setup.

	go run main.go microservice build-application-info --kube-config="/Users/freshteapot/.kube/config"
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			"git@github.com:freshteapot/test-deploy-key.git",
			"/tmp/dolittle-k8s",
			"auto-dev",
			// TODO fix this, then update deployment
			"/Users/freshteapot/dolittle/.ssh/test-deploy",
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
		nsName := ns.GetObjectMeta().GetName()
		if !strings.HasPrefix(nsName, "application-") {
			continue
		}

		annotationsMap := ns.GetObjectMeta().GetAnnotations()
		labelMap := ns.GetObjectMeta().GetLabels()

		applicationID := annotationsMap["dolittle.io/application-id"]
		tenantID := annotationsMap["dolittle.io/tenant-id"]

		application := platform.HttpResponseApplication{
			TenantID:     tenantID,
			ID:           applicationID,
			Name:         labelMap["application"],
			Environments: make([]platform.HttpInputEnvironment, 0),
		}

		namespace := nsName
		obj, err := client.AppsV1().Deployments(namespace).List(ctx, metaV1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		for _, item := range obj.Items {
			_, ok := item.ObjectMeta.Annotations["dolittle.io/tenant-id"]
			if !ok {
				continue
			}

			_, ok = item.ObjectMeta.Annotations["dolittle.io/application-id"]
			if !ok {
				continue
			}

			_, ok = item.ObjectMeta.Labels["application"]
			if !ok {
				continue
			}

			environment := platform.HttpInputEnvironment{
				Name:          item.ObjectMeta.Labels["environment"],
				TenantID:      tenantID,
				ApplicationID: applicationID,
			}

			index := funk.IndexOf(application.Environments, func(item platform.HttpInputEnvironment) bool {
				return item.Name == environment.Name
			})

			if index != -1 {
				continue
			}

			application.Environments = append(application.Environments, environment)
		}
		// Unique the environments
		applications = append(applications, application)
	}

	return applications
}

func init() {
	RootCmd.AddCommand(buildApplicationsCMD)
	buildApplicationsCMD.Flags().String("kube-config", "", "FullPath to kubeconfig")
}
