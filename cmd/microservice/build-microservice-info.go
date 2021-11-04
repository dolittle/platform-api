package microservice

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	gitStorage "github.com/dolittle-entropy/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type microserviceInfo struct {
	CustomerID     string
	ApplicationID  string
	Environment    string
	MicroserviceID string
	Name           string
	Kind           string
	Headimage      string
	Runtimeimage   string
	Ingress        []ingressInfo
}

type ingressInfo struct {
	Host         string
	Domainprefix string
	Path         string
	Pathtype     string
}

var buildMicroserviceInfoCMD = &cobra.Command{
	Use:   "build-microservice-info [CUSTOMERID] [APPLICATIONID] [ENVIRONMENT] [MICROSERVICEID] [FLAGS]",
	Short: "Resets the specified customers microservice configuration",
	Long: `
	It will attempt to update the git repo with resetted microservice configurations (ms_<microserviceID>.json).

	GIT_REPO_SSH_KEY="/Users/freshteapot/dolittle/.ssh/test-deploy" \
	GIT_REPO_BRANCH=dev \
	GIT_REPO_URL="git@github.com:freshteapot/test-deploy-key.git" \
	go run main.go microservice build-microservice-info --kube-config="/Users/freshteapot/.kube/config"
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
		kubeconfig := viper.GetString("tools.server.kubeConfig")

		if kubeconfig == "incluster" {
			kubeconfig = ""
		}

		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}

		// create the clientset
		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		createAll := viper.GetBool("all")
		shouldOverwrite := viper.GetBool("overwrite")

		if len(args) > 0 && createAll {
			logContext.Fatal("specify either the ID's or '--all' flag")
		}

		var microservices []microserviceInfo

		if createAll {
			logContext.Info("Discovering all microservices from the platform")
			// microservices = extractMicroservices(ctx, client)
		}

		switch len(args) {
		case 1:
			microservices = extractCustomersMicroservices(ctx, client, args[0], logContext)
		case 2:
			// microservices = extractApplicationsMicroservices(ctx, client, args[0], args[1], logContext)
		case 3:
			// microservices = extractEnvironmentsMicroservices(ctx, client, args[0], args[1], args[2], logContext)
		case 4:
			// microservices = extractMicroservice(ctx, client, args[0], args[1], args[2], args[3], logContext)
		default:
			logContext.Fatal("You have to specify some or all of the CUSTOMERID APPLICATIONID ENVIRONMENT MICROSERVICEID")
		}

		logContext.Infof("Creating microservice configurations for microservices: %v", microservices)
		CreateMicroserviceConfigs(gitRepo, microservices, shouldOverwrite, logContext)
		logContext.Info("Done!")
	},
}

// CreateMicroserviceConfigs resets all of the found customers studio.json files to enable automation for all environments
// and to enable overwriting
func CreateMicroserviceConfigs(repo storage.Repo, microservices []microserviceInfo, shouldOverwrite bool, logger logrus.FieldLogger) error {
	logContext := logger.WithFields(logrus.Fields{
		"function": "CreateMicroserviceConfigs",
	})

	for _, microservice := range microservices {
		if err := repo.SaveMicroservice(microservice.CustomerID, microservice.ApplicationID, microservice.Environment, microservice.MicroserviceID, microservice); err != nil {
			logContext.WithFields(logrus.Fields{
				"error":          err,
				"customerID":     microservice.CustomerID,
				"applicationID":  microservice.ApplicationID,
				"environment":    microservice.Environment,
				"microserviceID": microservice.MicroserviceID,
			}).Fatal("couldn't save microservice config")
		}
	}
	return nil
}

func extractCustomersMicroservices(ctx context.Context, client kubernetes.Interface, customerID string, logger logrus.FieldLogger) []microserviceInfo {
	logContext := logger.WithFields(logrus.Fields{
		"function":   "extractCustomersMicroservices",
		"customerID": customerID,
	})
	customerID = strings.ToLower(customerID)

	var microservices []microserviceInfo
	for _, namespace := range getNamespaces(ctx, client) {
		if isApplicationNamespace(namespace) {
			if namespace.Annotations["dolittle.io/tenant=id"] == customerID {
				deployments, err := client.AppsV1().Deployments(namespace.GetName()).List(ctx, v1.ListOptions{})
				if err != nil {
					logContext.WithFields(logrus.Fields{
						"error":     err,
						"namespace": namespace.GetName(),
					}).Fatalf("couldn't get the namespaces deployments")
				}

				for _, deployment := range deployments.Items {
					microserviceID, hasAnnotation := deployment.Annotations["dolittle.io/microservice-id"]
					if !hasAnnotation {
						continue
					}

					applicationID := deployment.Annotations["dolittle.io/application-id"]
					environment := deployment.Labels["environment"]
					kind, hasAnnotation := deployment.Annotations["dolittle.io/microservice-kind"]
					if !hasAnnotation {
						kind = "unknown"
					}
					name := deployment.GetName()

					var head, runtime string

					// get images
					for _, container := range deployment.Spec.Template.Spec.Containers {
						switch container.Name {
						case "head":
							head = container.Image
						case "runtime":
							runtime = container.Image
						default:
							continue
						}
					}

					var ingresses []ingressInfo

					selector := fmt.Sprintf("microservice=%s", deployment.Labels["microservice"])
					foundIngresses, err := client.NetworkingV1beta1().Ingresses(namespace.GetName()).List(ctx, v1.ListOptions{
						LabelSelector: selector,
					})

					if err != nil {
						logContext.WithFields(logrus.Fields{
							"error":     err,
							"namespace": namespace.GetName(),
							"selector":  selector,
						}).Fatalf("couldn't get the namespaces ingresses")
					}

					for _, ingress := range foundIngresses.Items {
						for _, ingressRule := range ingress.Spec.Rules {
							for _, httpRule := range ingressRule.HTTP.Paths {

								ingressInfo := ingressInfo{
									Host:         ingressRule.Host,
									Domainprefix: strings.TrimSuffix(ingressRule.Host, ".dolittle.cloud"),
									Path:         httpRule.Path,
									Pathtype:     string(*httpRule.PathType),
								}
								ingresses = append(ingresses, ingressInfo)
							}
						}
					}

					microserviceInfo := microserviceInfo{
						CustomerID:     customerID,
						ApplicationID:  applicationID,
						Environment:    strings.ToLower(environment),
						MicroserviceID: microserviceID,
						Name:           name,
						Kind:           kind,
						Headimage:      head,
						Runtimeimage:   runtime,
					}

					microservices = append(microservices, microserviceInfo)
				}
			}
		}
	}
	return microservices
}

func init() {
	RootCmd.AddCommand(buildMicroserviceInfoCMD)

	buildMicroserviceInfoCMD.Flags().Bool("all", false, "Discovers all of the microservices from the platform")
	viper.BindPFlag("all", buildMicroserviceInfoCMD.Flags().Lookup("all"))

	buildMicroserviceInfoCMD.Flags().Bool("overwrite", false, "Overwrites existing microservice configuration files")
	viper.BindPFlag("overwrite", buildMicroserviceInfoCMD.Flags().Lookup("overwrite"))
}
