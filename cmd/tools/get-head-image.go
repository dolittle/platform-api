package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/dolittle/platform-api/pkg/git"
	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thoas/go-funk"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sJson "k8s.io/apimachinery/pkg/runtime/serializer/json"
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
		dryRun, _ := cmd.Flags().GetBool("dry-run")

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

		stream, _ := ioutil.ReadFile(filePath)
		rawResources, err := SplitYAML(stream)
		if err != nil {
			fmt.Println("after splitting the yaml")
			panic(err)
		}

		currentIndex, currentDeployment, rawYAML, err := getDeploymentFromMicroserviceYAML(rawResources)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"error":         err,
				"file_path":     filePath,
				"current_index": currentIndex,
			}).Error("failed getDeploymentFromMicroserviceYAML")
			return
		}

		updateImages(deployment, currentDeployment, rawYAML)

		s := runtime.NewScheme()
		serializer := k8sJson.NewSerializerWithOptions(
			k8sJson.DefaultMetaFactory,
			s,
			s,
			k8sJson.SerializerOptions{
				Yaml:   true,
				Pretty: true,
				Strict: true,
			},
		)

		if dryRun {
			serializer.Encode(&currentDeployment, os.Stdout)
			return
		}

		outputFilePath := fmt.Sprintf("%s/Source/V3/Kubernetes/Customers/%s/%s/%s/%s/microservice-deployment.yml",
			gitRepoConfig.RepoRoot,
			customer,
			applicationName,
			environment,
			microserviceName,
		)
		file, _ := os.Create(outputFilePath)
		defer file.Close()
		serializer.Encode(&currentDeployment, file)

		type Summary struct {
			FileName string `json:"file_name"`
		}

		summary := Summary{
			FileName: outputFilePath,
		}

		b, _ := json.Marshal(summary)
		fmt.Println(string(b))
	},
}

func init() {
	git.SetupViper()
	getHeadImageCMD.Flags().String("application-id", "", "Name of Application")
	getHeadImageCMD.Flags().String("environment", "prod", "Application environment")
	getHeadImageCMD.Flags().String("microservice-id", "", "microservice-id")
	getHeadImageCMD.Flags().Bool("dry-run", false, "Output the data, do not write to disk")
}

func updateContainerImage(rawYAML yaml.Node, name string, image string) {
	p, err := yamlpath.NewPath(
		fmt.Sprintf(`$..spec.containers[*][?(@.name=="%s")].image`, name),
	)

	if err != nil {
		log.Fatalf("cannot create path: %v", err)
	}

	q, err := p.Find(&rawYAML)
	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}

	q[0].Value = image
}

func updateImages(cluster v1.Deployment, current v1.Deployment, rawYAML yaml.Node) {
	for _, clusterContainer := range cluster.Spec.Template.Spec.Containers {
		index := funk.IndexOf(current.Spec.Template.Spec.Containers, func(container corev1.Container) bool {
			return clusterContainer.Name == container.Name
		})

		if index == -1 {
			fmt.Println("Not found, might be a bad thing")
			continue
		}

		currentContainer := current.Spec.Template.Spec.Containers[index]

		if clusterContainer.Image != currentContainer.Image {
			// This is what we need if we drop the yaml stuff
			current.Spec.Template.Spec.Containers[index].Image = clusterContainer.Image

			updateContainerImage(rawYAML, clusterContainer.Name, clusterContainer.Image)
		}
	}
}

func SplitYAML(resources []byte) ([][]byte, error) {
	dec := yaml.NewDecoder(bytes.NewReader(resources))

	var res [][]byte
	for {
		var value interface{}
		err := dec.Decode(&value)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		valueBytes, err := yaml.Marshal(value)
		if err != nil {
			return nil, err
		}
		res = append(res, valueBytes)
	}
	return res, nil
}

func getDeploymentFromMicroserviceYAML(resources [][]byte) (index int, deployment v1.Deployment, rawYAML yaml.Node, err error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode

	for index, resource := range resources {
		var n yaml.Node
		err := yaml.Unmarshal(resource, &n)
		if err != nil {
			log.Fatalf("cannot unmarshal data: %v", err)
		}

		obj, gKV, err := decode(resource, nil, nil)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if gKV.Kind == "Deployment" {
			yaml.Unmarshal(resource, &rawYAML)
			deployment := obj.(*v1.Deployment)
			return index, *deployment, rawYAML, nil
		}
	}
	return -1, v1.Deployment{}, rawYAML, errors.New("not-found")
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
