package microservice

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/dolittle/platform-api/pkg/git"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	coreV1 "k8s.io/api/core/v1"
	netV1 "k8s.io/api/networking/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var buildApplicationInfoCMD = &cobra.Command{
	Use:   "build-application-info",
	Short: "Write application info into the git repo",
	Long: `
	It will attempt to update the git repo with data from the cluster and skip those that have been setup.

	GIT_REPO_SSH_KEY="/Users/freshteapot/dolittle/.ssh/test-deploy" \
	GIT_REPO_BRANCH=dev \
	GIT_REPO_URL="git@github.com:freshteapot/test-deploy-key.git" \
	go run main.go microservice build-application-info --kube-config="/Users/freshteapot/.kube/config"
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logContext := logrus.StandardLogger()
		gitRepoConfig := git.InitGit(logContext)

		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		ctx := context.TODO()
		kubeconfig := viper.GetString("tools.server.kubeConfig")

		if kubeconfig == "incluster" {
			kubeconfig = ""
		}
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

		shouldCommit := viper.GetBool("commit")

		logContext.Info("Starting to extract applications from the cluster")
		applications := extractApplications(ctx, client)

		logContext.Info(fmt.Sprintf("Saving %v application(s)", len(applications)))
		SaveApplications(gitRepo, applications, shouldCommit, logContext)
		logContext.Info("Done!")
	},
}

// SaveApplications saves the Applications into applications.json and also creates a default studio.json if
// the customer doesn't have one
func SaveApplications(repo storage.Repo, applications []platform.HttpResponseApplication, shouldCommit bool, logger logrus.FieldLogger) error {
	logContext := logger.WithFields(logrus.Fields{
		"function": "SaveApplications",
	})
	for _, application := range applications {
		studioConfig, err := repo.GetStudioConfig(application.TenantID)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"error":      err,
				"customerID": application.TenantID,
			}).Fatalf("didn't find a studio config for customer %s, create a config for this customer by running 'microservice build-studio-info %s'",
				application.TenantID, application.TenantID)
		}

		if !studioConfig.BuildOverwrite {
			continue
		}
		if err = repo.SaveApplication(application); err != nil {
			logContext.WithFields(logrus.Fields{
				"error":    err,
				"tenantID": application.TenantID,
			}).Fatal("failed to save application")
		}
	}
	return nil
}

func extractApplications(ctx context.Context, client kubernetes.Interface) []platform.HttpResponseApplication {
	applications := make([]platform.HttpResponseApplication, 0)

	for _, namespace := range getNamespaces(ctx, client) {
		if isApplicationNamespace(namespace) {
			applications = append(applications, getApplicationFromK8s(ctx, client, namespace))
		}
	}

	return applications
}

func getNamespaces(ctx context.Context, client kubernetes.Interface) []coreV1.Namespace {
	namespacesList, err := client.CoreV1().Namespaces().List(ctx, metaV1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	return namespacesList.Items
}

func isApplicationNamespace(namespace coreV1.Namespace) bool {
	if !strings.HasPrefix(namespace.GetName(), "application-") {
		return false
	}
	if _, hasAnnotation := namespace.Annotations["dolittle.io/tenant-id"]; !hasAnnotation {
		return false
	}
	if _, hasAnnotation := namespace.Annotations["dolittle.io/application-id"]; !hasAnnotation {
		return false
	}
	if _, hasLabel := namespace.Labels["tenant"]; !hasLabel {
		return false
	}
	if _, hasLabel := namespace.Labels["application"]; !hasLabel {
		return false
	}

	return true
}

func getApplicationFromK8s(ctx context.Context, client kubernetes.Interface, namespace coreV1.Namespace) platform.HttpResponseApplication {
	application := platform.HttpResponseApplication{
		ID:         namespace.Annotations["dolittle.io/application-id"],
		Name:       namespace.Labels["application"],
		TenantID:   namespace.Annotations["dolittle.io/tenant-id"],
		TenantName: namespace.Labels["tenant"],
	}

	application.Environments = getApplicationEnvironmentsFromK8s(ctx, client, namespace.GetName(), application.ID, application.TenantID)

	return application
}

func getApplicationEnvironmentsFromK8s(ctx context.Context, client kubernetes.Interface, namespace, applicationID, tenantID string) []platform.HttpInputEnvironment {
	environments := make([]platform.HttpInputEnvironment, 0)
	for _, configmap := range getConfigmaps(ctx, client, namespace) {
		if isEnvironmentTenantsConfig(configmap) {
			environment := platform.HttpInputEnvironment{
				Name:          configmap.Labels["environment"],
				TenantID:      tenantID,
				ApplicationID: applicationID,
			}

			environmentLabels := configmap.Labels

			environment.Tenants = getTenantsFromTenantsJson(configmap.Data["tenants.json"])
			environment.Ingresses = getEnvironmentIngressesFromK8s(ctx, client, namespace, environmentLabels)

			environments = append(environments, environment)
		}
	}
	return environments
}

func getConfigmaps(ctx context.Context, client kubernetes.Interface, namespace string) []coreV1.ConfigMap {
	configmapList, err := client.CoreV1().ConfigMaps(namespace).List(ctx, metaV1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	return configmapList.Items
}

func isEnvironmentTenantsConfig(configmap coreV1.ConfigMap) bool {
	if _, hasAnnotation := configmap.Annotations["dolittle.io/tenant-id"]; !hasAnnotation {
		return false
	}
	if _, hasAnnotation := configmap.Annotations["dolittle.io/application-id"]; !hasAnnotation {
		return false
	}
	if _, hasLabel := configmap.Labels["tenant"]; !hasLabel {
		return false
	}
	if _, hasLabel := configmap.Labels["application"]; !hasLabel {
		return false
	}
	if _, hasLabel := configmap.Labels["environment"]; !hasLabel {
		return false
	}

	_, hasTenantsJson := configmap.Data["tenants.json"]
	return hasTenantsJson
}

func getTenantsFromTenantsJson(tenantsJsonContent string) []platform.TenantId {
	tenantObjects := make(map[string]interface{})
	if err := json.Unmarshal([]byte(tenantsJsonContent), &tenantObjects); err != nil {
		panic(err.Error())
	}

	tenants := make([]platform.TenantId, 0)
	for tenantID := range tenantObjects {
		tenants = append(tenants, platform.TenantId(tenantID))
	}
	return tenants
}

func getEnvironmentIngressesFromK8s(ctx context.Context, client kubernetes.Interface, namespace string, environmentLabels labels.Set) platform.EnvironmentIngresses {
	ingresses := make(map[platform.TenantId]platform.EnvironmentIngress, 0)
	for _, ingress := range getIngresses(ctx, client, namespace, environmentLabels) {
		if !isMicroserviceIngress(ingress) {
			continue
		}

		tenantIDFound, tenantID := tryGetTenantFromIngress(ingress)
		if !tenantIDFound {
			continue
		}

		for _, rule := range ingress.Spec.Rules {
			host := rule.Host
			domainPrefix := strings.TrimSuffix(host, ".dolittle.cloud")

			secretNameFound, secretName := tryGetIngressSecretNameForHost(ingress, host)

			if secretNameFound {
				environmentIngress := platform.EnvironmentIngress{
					Host:         host,
					DomainPrefix: domainPrefix,
					SecretName:   secretName,
				}

				ingresses[tenantID] = environmentIngress
				break
			}
		}
	}
	return ingresses
}

func getIngresses(ctx context.Context, client kubernetes.Interface, namespace string, environmentLabels labels.Set) []netV1.Ingress {
	ingressList, err := client.NetworkingV1().Ingresses(namespace).List(ctx, metaV1.ListOptions{
		LabelSelector: labels.FormatLabels(environmentLabels),
	})
	if err != nil {
		panic(err.Error())
	}
	return ingressList.Items
}

func isMicroserviceIngress(ingress netV1.Ingress) bool {
	if _, hasAnnotation := ingress.Annotations["dolittle.io/tenant-id"]; !hasAnnotation {
		return false
	}
	if _, hasAnnotation := ingress.Annotations["dolittle.io/application-id"]; !hasAnnotation {
		return false
	}
	if _, hasAnnotation := ingress.Annotations["dolittle.io/microservice-id"]; !hasAnnotation {
		return false
	}
	if _, hasLabel := ingress.Labels["tenant"]; !hasLabel {
		return false
	}
	if _, hasLabel := ingress.Labels["application"]; !hasLabel {
		return false
	}
	if _, hasLabel := ingress.Labels["environment"]; !hasLabel {
		return false
	}
	if _, hasLabel := ingress.Labels["microservice"]; !hasLabel {
		return false
	}

	return true
}

var tenantFromIngressAnnotationExtractor = regexp.MustCompile(`proxy_set_header\s+Tenant-ID\s+"([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})"`)

func tryGetTenantFromIngress(ingress netV1.Ingress) (bool, platform.TenantId) {
	tenantHeaderAnnotation := ingress.GetObjectMeta().GetAnnotations()["nginx.ingress.kubernetes.io/configuration-snippet"]
	tenantID := tenantFromIngressAnnotationExtractor.FindStringSubmatch(tenantHeaderAnnotation)
	if tenantID == nil {
		return false, ""
	}
	return true, platform.TenantId(tenantID[1])
}

func tryGetIngressSecretNameForHost(ingress netV1.Ingress, host string) (bool, string) {
	for _, tlsConfig := range ingress.Spec.TLS {
		for _, tlsHost := range tlsConfig.Hosts {
			if tlsHost == host {
				return true, tlsConfig.SecretName
			}
		}
	}
	return false, ""
}

func init() {
	RootCmd.AddCommand(buildApplicationInfoCMD)
}
