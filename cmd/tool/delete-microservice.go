package tool

import (
	"context"
	"fmt"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	gitHelper "github.com/dolittle-entropy/platform-api/pkg/platform/git"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/purchaseorderapi"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/rawdatalog"
	gitStorage "github.com/dolittle-entropy/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var deleteMicroserviceCMD = &cobra.Command{
	Use:   "delete-microservice",
	Short: "Delete a microservice from Studio and the platform",
	Long: `

	go run main.go tool delete-microservice \
	--kube-config="/Users/freshteapot/.kube/config" \
	--tenant-id=ID \
	--microservice-id=ID \
	--application-id=ID \
	--environment=PLATFORM_ENV \
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO Delete today does not include environment
		logrus.SetFormatter(&logrus.JSONFormatter{})

		var (
			err error
		)

		ctx := context.TODO()
		kubeconfig, _ := cmd.Flags().GetString("kube-config")

		tenantID, _ := cmd.Flags().GetString("tenant-id")
		applicationID, _ := cmd.Flags().GetString("application-id")
		environment, _ := cmd.Flags().GetString("environment")
		microserviceID, _ := cmd.Flags().GetString("microservice-id")

		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}

		// create the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		fmt.Println(ctx, config)

		logContext := logrus.WithFields(logrus.Fields{
			"tenant_id":       tenantID,
			"application_id":  applicationID,
			"environment":     environment,
			"microservice_id": microserviceID,
		})

		logContext.Info("Delete microservice")
		return
		gitRepoConfig := gitHelper.InitGit(logContext)
		k8sRepo := platform.NewK8sRepo(clientset, config)

		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		simpleRepo := microservice.NewSimpleRepo(clientset)
		businessMomentsAdaptorRepo := microservice.NewBusinessMomentsAdaptorRepo(clientset)

		rawDataLogRepo := rawdatalog.NewRawDataLogIngestorRepo(k8sRepo, clientset, gitRepo, logContext)
		specFactory := purchaseorderapi.NewK8sResourceSpecFactory()
		k8sResources := purchaseorderapi.NewK8sResource(clientset, specFactory)
		purchaseOrderApiRepo := purchaseorderapi.NewRepo(k8sResources, specFactory, clientset)

		gitRepo.DeleteMicroservice(tenantID, applicationID, environment, microserviceID)

		namespace := fmt.Sprintf("application-%s", applicationID)

		err = simpleRepo.Delete(namespace, microserviceID)
		if err != nil {
			fmt.Println(err)
		}

		err = businessMomentsAdaptorRepo.Delete(namespace, microserviceID)
		if err != nil {
			fmt.Println(err)
		}

		err = rawDataLogRepo.Delete(namespace, microserviceID)
		if err != nil {
			fmt.Println(err)
		}

		err = purchaseOrderApiRepo.Delete(namespace, microserviceID)
		if err != nil {
			fmt.Println(err)
		}
	},
}
