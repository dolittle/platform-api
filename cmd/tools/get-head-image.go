package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/dolittle/platform-api/pkg/git"
	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thoas/go-funk"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var getHeadImageCMD = &cobra.Command{
	Use:   "get-head-image",
	Short: "Write terraform output for a customer",
	Long: `
	Outputs a new Dolittle platform customer in hcl to stdout.

	go run main.go tools get-head-image --application-id="" --environment="" --microservice-id=""
	`,
	Run: func(cmd *cobra.Command, args []string) {

		// Lookup image based on microserviceID?
		// Lookup image based on application/env/microserviceID?
		// Get all
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logContext := logrus.StandardLogger()
		gitRepoConfig := git.InitGit(logContext)

		_ = gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		// fix: https://github.com/spf13/viper/issues/798
		for _, key := range viper.AllKeys() {
			viper.Set(key, viper.Get(key))
		}

		config, err := getKubeRestConfig(viper.GetString("tools.server.kubeConfig"))
		if err != nil {
			panic(err.Error())
		}

		// create the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		applicationID, _ := cmd.Flags().GetString("application-id")
		microserviceID, _ := cmd.Flags().GetString("microservice-id")
		environment, _ := cmd.Flags().GetString("environment")

		deployment := getDeployment(clientset, applicationID, environment, microserviceID)

		if deployment.GetObjectMeta().GetName() == "" {
			logContext.Error("Microservice not found")
			return
		}
		// TODO check if in platform?
		// Lookup in manual location
		customer := deployment.Labels["tenant"]
		applicationName := deployment.Labels["application"]
		microserviceName := deployment.Labels["microservice"]

		filePath := fmt.Sprintf("%s/Source/V3/Kubernetes/Customers/%s/%s/%s/%s/microservice.yml",
			gitRepoConfig.RepoRoot,
			customer,
			applicationName,
			environment,
			microserviceName,
		)
		// TODO use currentIndex for saving
		_, currentDeployment, err := getDeploymentFromMicroserviceYAML(filePath)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"error":     err,
				"file_path": filePath,
			}).Error("failed getDeploymentFromMicroserviceYAML")
			return
		}

		getOutput(applicationID, environment, microserviceID, deployment, currentDeployment)
	},
}

func init() {
	git.SetupViper()
	getHeadImageCMD.Flags().String("application-id", "", "Name of Application")
	getHeadImageCMD.Flags().String("environment", "prod", "Application environment")
	getHeadImageCMD.Flags().String("microservice-id", "", "microservice-id")
}

func getOutput(applicationID string, environment string, microserviceID string, cluster v1.Deployment, current v1.Deployment) {
	//tenantID := found.ObjectMeta.Annotations["dolittle.io/tenant-id"]
	//gitRepo.GetMicroservice(tenantID, applicationID, environment, microserviceID)
	type output struct {
		Name           string `json:"name"`
		CurrentImage   string `json:"current_name"`
		ClusterImage   string `json:"cluster_name"`
		Match          bool   `json:"match"`
		ApplicationID  string `json:"application_id"`
		MicroserviceID string `json:"microservice_id"`
		Environment    string `json:"environment"`
	}

	for _, clusterContainer := range cluster.Spec.Template.Spec.Containers {
		found := funk.Find(current.Spec.Template.Spec.Containers, func(container corev1.Container) bool {
			return clusterContainer.Name == container.Name
		})

		if found == nil {
			//fmt.Println("skip")
			continue
		}

		current := found.(corev1.Container)
		info := output{
			Name:           current.Name,
			CurrentImage:   current.Image,
			ClusterImage:   clusterContainer.Image,
			ApplicationID:  applicationID,
			MicroserviceID: microserviceID,
			Environment:    environment,
		}

		info.Match = info.ClusterImage == info.CurrentImage
		b, _ := json.Marshal(info)
		fmt.Println(string(b))
	}
}

func getDeploymentFromMicroserviceYAML(filePath string) (position int, deployment v1.Deployment, err error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	stream, _ := ioutil.ReadFile(filePath)

	data := string(stream)
	parts := strings.Split(data, "---\n")
	for index, part := range parts {
		b := []byte(part)
		obj, gKV, err := decode(b, nil, nil)
		if err != nil {
			continue
		}

		if gKV.Kind == "Deployment" {
			deployment := obj.(*v1.Deployment)
			// Maybe we can use DeepCopy
			return index, *deployment, nil
		}
	}
	return -1, v1.Deployment{}, errors.New("not-found")
}

func getDeployment(k8sClient kubernetes.Interface, applicationID string, environment string, microserviceID string) v1.Deployment {
	namespace := fmt.Sprintf("application-%s", applicationID)
	ctx := context.TODO()
	options := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("environment=%s", environment),
	}

	deployments, _ := k8sClient.AppsV1().Deployments(namespace).List(ctx, options)
	var found v1.Deployment
	for _, deployment := range deployments.Items {
		annotations := deployment.GetAnnotations()

		// the microserviceID is unique per microservice so that's enough for the check
		if annotations["dolittle.io/microservice-id"] == microserviceID {
			found = deployment
			break
		}
	}
	return found
}

// TODO move this and share
func getKubeRestConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig == "incluster" {
		kubeconfig = ""
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}
