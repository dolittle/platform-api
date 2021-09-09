package microservice

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	gitStorage "github.com/dolittle-entropy/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/thoas/go-funk"
	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	netV1 "k8s.io/api/networking/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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

// SaveApplications saves the Applications
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

	for _, ns := range getNamespaces(ctx, client) {
		// Creates a single application for a namespace(?)
		applications = append(applications, getApplicationFromK8s(ctx, client, ns))
	}

	return applications
}

func getNamespaces(ctx context.Context, client *kubernetes.Clientset) []v1.Namespace {
	namespacesList, err := client.CoreV1().Namespaces().List(ctx, metaV1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	return namespacesList.Items
}

func getApplicationFromK8s(ctx context.Context, client *kubernetes.Clientset, namespace v1.Namespace) platform.HttpResponseApplication {
	application := createBasicApplicationStructureFromMetadata(client, namespace)
	addEnvironmentsFromK8sInto(&application, ctx, client, namespace.GetObjectMeta().GetName())
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

func addEnvironmentsFromK8sInto(application *platform.HttpResponseApplication, ctx context.Context, client *kubernetes.Clientset, namespaceName string) {
	for _, deployment := range getDeployments(ctx, client, namespaceName) {
		if !deploymentHasRequiredMetadata(deployment) {
			continue
		}

		environment := createOrGetEnvironment(ctx, client, deployment, namespaceName, *application)
		addIngressesIntoEnvironment(&environment, ctx, client, namespaceName, deployment)

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

func createOrGetEnvironment(ctx context.Context, client *kubernetes.Clientset, deployment appsV1.Deployment, namespace string, application platform.HttpResponseApplication) platform.HttpInputEnvironment {
	environmentName := deployment.ObjectMeta.Labels["environment"]
	environmentExists, i := environmentAlreadyInApplication(application.Environments, environmentName)
	if environmentExists {
		return application.Environments[i]
	}
	return platform.HttpInputEnvironment{
		Name:          environmentName,
		TenantID:      application.TenantID,
		ApplicationID: application.ID,
		Tenants:       getTenants(ctx, client, namespace, environmentName),
		Ingresses:     make(map[platform.TenantId]platform.EnvironmentIngress),
	}
}

func getTenants(ctx context.Context, client *kubernetes.Clientset, namespace, environmentName string) []platform.TenantId {
	tenants := make([]platform.TenantId, 0)

	for tenantStr, _ := range getTenantsFromConfigmap(ctx, client, namespace, environmentName) {
		tenants = append(tenants, platform.TenantId(tenantStr))
	}
	return tenants
}

func addIngressesIntoEnvironment(environment *platform.HttpInputEnvironment, ctx context.Context, client *kubernetes.Clientset, namespace string, deployment appsV1.Deployment) map[platform.TenantId]platform.EnvironmentIngress {
	ingresses := environment.Ingresses
	for _, ingress := range getIngresses(ctx, client, namespace, deployment) {
		if len(ingress.Spec.TLS) > 0 {
			// Assume that there is only one Rule, or that only the top one counts
			// Note that this will override ingresses for tenants, not a problem if they are the same, but confusing if there are different configs
			ingresses[getTenantFromIngress(ingress)] = extractEnvironmentIngressFromIngressRule(ingress.Spec.Rules[0])
		}
	}
	return ingresses
}

func extractEnvironmentIngressFromIngressRule(rule netV1.IngressRule) platform.EnvironmentIngress {
	host := rule.Host
	domainPrefix := strings.ReplaceAll(host, ".dolittle.cloud", "")

	// Note that this will override ingresses for tenants, not a problem if they are the same, but confusing if there are different configs
	return platform.EnvironmentIngress{
		Host:         host,
		DomainPrefix: domainPrefix,
	}
}
func getTenantFromIngress(ingress netV1.Ingress) platform.TenantId {
	tenantHeaderAnnotation := ingress.GetObjectMeta().GetAnnotations()["nginx.ingress.kubernetes.io/configuration-snippet"]
	tenantID := strings.ReplaceAll(tenantHeaderAnnotation, "proxy_set_header Tenant-ID", "")
	tenantID = strings.ReplaceAll(tenantID, "\"", "")
	return platform.TenantId(strings.TrimSpace(tenantID))
}

func getIngresses(ctx context.Context, client *kubernetes.Clientset, namespace string, deployment appsV1.Deployment) []netV1.Ingress {
	ingresses, err := client.NetworkingV1().Ingresses(namespace).List(ctx, metaV1.ListOptions{
		LabelSelector: labels.FormatLabels(deployment.Labels),
	})
	if err != nil {
		panic(err.Error())
	}
	return ingresses.Items
}

func getTenantsFromConfigmap(ctx context.Context, client *kubernetes.Clientset, namespace, environmentName string) map[string]interface{} {
	tenantsConfig, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, fmt.Sprintf("%s-tenants", strings.ToLower(environmentName)), metaV1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	tenantsJson := tenantsConfig.Data["tenants.json"]
	tenants := make(map[string]interface{})
	err = json.Unmarshal([]byte(tenantsJson), &tenants)
	if err != nil {
		panic(err.Error())
	}
	return tenants
}

func deploymentHasRequiredMetadata(deployment appsV1.Deployment) bool {
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

func environmentAlreadyInApplication(environments []platform.HttpInputEnvironment, environmentName string) (bool, int) {
	index := funk.IndexOf(environments, func(item platform.HttpInputEnvironment) bool {
		return item.Name == environmentName
	})

	if index != -1 {
		return false, index
	}
	return true, index
}

func init() {
	RootCmd.AddCommand(buildApplicationsCMD)
	buildApplicationsCMD.Flags().String("kube-config", "", "FullPath to kubeconfig")
}
