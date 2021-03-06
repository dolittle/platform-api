package upsert

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dolittle/platform-api/pkg/git"
	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform/automate"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/manual"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var applicationCMD = &cobra.Command{
	Use:   "application",
	Short: "Upsert the application.json",
	Long: `
	Will upsert the application.json in the git repo with data from the cluster and skip those that have been setup.

	GIT_REPO_BRANCH=dev \
	GIT_REPO_DRY_RUN=true \
	GIT_REPO_DIRECTORY="/tmp/dolittle-local-dev" \
	GIT_REPO_DIRECTORY_ONLY=true \
	go run main.go tools studio upsert application
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)
		logger := logrus.StandardLogger()
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		resetAll, _ := cmd.Flags().GetBool("all")

		logContext := logger.WithFields(logrus.Fields{
			"cmd": "tools studio upsert application",
		})

		platformEnvironment := viper.GetString("tools.server.platformEnvironment")
		gitRepoConfig := git.InitGit(logContext, platformEnvironment)

		storageRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		k8sClient, _ := platformK8s.InitKubernetesClient()
		k8sRepoV2 := k8s.NewRepo(k8sClient, logger.WithField("context", "k8s-repo-v2"))
		manualRepo := manual.NewManualHelper(k8sClient, k8sRepoV2, storageRepo, logContext.WithField("context", "manual-repo"))

		oneNamespace := ""
		if len(args) > 0 {
			oneNamespace = args[0]
		}

		if oneNamespace != "" && resetAll {
			logContext.Fatal("specify either a namespace or '--all' flag")
		}

		var resources []corev1.Namespace
		if resetAll {
			logContext.Info("Discovering all namespaces with an application in the platform")
			resources, _ = k8sRepoV2.GetNamespacesWithApplication()
		}

		if oneNamespace != "" {

			ctx := context.TODO()
			resource, err := k8sClient.CoreV1().Namespaces().Get(ctx, oneNamespace, metav1.GetOptions{})
			if err != nil {
				logContext.WithFields(logrus.Fields{
					"namesapce": oneNamespace,
					"error":     err,
				}).Fatal("failed to lookup namespace")
			}
			resources = append(resources, *resource)
		}

		if len(resources) == 0 {
			logContext.Fatal("No namespaces found")
		}

		for _, resource := range resources {
			if !automate.IsApplicationNamespace(resource) {
				continue
			}

			namespace := resource.Name
			customerID := resource.Annotations["dolittle.io/tenant-id"]
			applicationID := resource.Annotations["dolittle.io/application-id"]
			logContext := logger.WithFields(logrus.Fields{
				"customer_id":    customerID,
				"application_id": applicationID,
			})

			processOne(storageRepo, platformEnvironment, manualRepo, namespace, dryRun, logContext)
		}

	},
}

func processOne(
	storageRepo storage.Repo,
	platformEnvironment string,
	manualRepo manual.Repo,
	namespace string,
	dryRun bool,
	logContext logrus.FieldLogger,
) {
	application, err := manualRepo.GatherOne(platformEnvironment, namespace)

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to find namespace")
		return
	}

	logContext = logContext.WithFields(logrus.Fields{
		"application_id": application.ID,
	})

	customer, err := storageRepo.GetTerraformTenant(application.CustomerID)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error":       err,
			"customer_id": application.CustomerID,
			"tip": fmt.Sprintf(
				"didn't find a studio config for customer %s, create a config for this customer by running 'tools studio build-studio-info %s'",
				application.CustomerID,
				application.CustomerID,
			),
		}).Error("Failed to find customer terraform")
		return

	}

	if customer.PlatformEnvironment != platformEnvironment {
		logContext.WithFields(logrus.Fields{
			"error":                         "platform-environment",
			"customer_platform_environment": customer.PlatformEnvironment,
			"platform_environment":          platformEnvironment,
		}).Warn("Skipping")
		return
	}

	studioConfig, err := storageRepo.GetStudioConfig(application.CustomerID)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error":       err,
			"customer_id": application.CustomerID,
			"tip": fmt.Sprintf(
				"didn't find a studio config for customer %s, create a config for this customer by running 'tools studio build-studio-info %s'",
				application.CustomerID,
				application.CustomerID,
			),
		}).Error("Failed to find studio config")
		return
	}

	if !studioConfig.BuildOverwrite {
		logContext.WithFields(logrus.Fields{
			"error":       "build-overwrite-set",
			"customer_id": application.CustomerID,
		}).Warn("Skipping")
		return
	}

	// Check what is there?
	currentApplicaiton, err := storageRepo.GetApplication(application.CustomerID, application.ID)
	if err != nil {
		if err != storage.ErrNotFound {
			logContext.WithFields(logrus.Fields{
				"error":     err,
				"namespace": namespace,
			}).Fatal("Failed to get application")
		}
	}

	// Keep track of welcomeMicroserviceID
	// Get currentEnvironmentWelcomeMicroservices
	currentEnvironmentWelcomeMicroservices := make(map[string]string)
	for _, environment := range currentApplicaiton.Environments {
		currentEnvironmentWelcomeMicroservices[environment.Name] = environment.WelcomeMicroserviceID
	}

	for index, environment := range application.Environments {
		if welcomeMicroserviceID, ok := currentEnvironmentWelcomeMicroservices[environment.Name]; ok {
			environment.WelcomeMicroserviceID = welcomeMicroserviceID
			application.Environments[index] = environment
		}
	}

	// Set the status from what is already there
	application.Status = currentApplicaiton.Status

	if dryRun {
		b, _ := json.Marshal(application)
		fmt.Println(string(b))
		return
	}

	err = storageRepo.SaveApplication(application)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error":     err,
			"namespace": namespace,
		}).Fatal("Failed to write applicaiton")
	}
}

func init() {
	applicationCMD.Flags().Bool("dry-run", true, "Will not write to disk")
	applicationCMD.Flags().Bool("all", false, "Discovers all customers from the platform and resets all studio.json's to default state (everything allowed)")
}
