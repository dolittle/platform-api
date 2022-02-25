package upsert

import (
	"os"

	"github.com/dolittle/platform-api/pkg/git"
	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var studioCMD = &cobra.Command{
	Use:   "studio [CUSTOMERID]... [FLAGS]",
	Short: "Upsert the specified customers studio configuration",
	Long: `
	Will update the git repo with resetted Studio configurations (studio.json).

	GIT_REPO_BRANCH=dev \
	GIT_REPO_DRY_RUN=true \
	GIT_REPO_DIRECTORY="/tmp/dolittle-local-dev" \
	GIT_REPO_DIRECTORY_ONLY=true \
	go run main.go tools studio upsert studio
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logger := logrus.StandardLogger()
		resetAll, _ := cmd.Flags().GetBool("all")
		disabledEnvironments, _ := cmd.Flags().GetBool("disable-environments")
		disableCanCreateApplication, _ := cmd.Flags().GetBool("disable-create-application")

		platformEnvironment := viper.GetString("tools.server.platformEnvironment")
		gitRepoConfig := git.InitGit(logger, platformEnvironment)

		gitRepo := gitStorage.NewGitStorage(
			logger.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		logContext := logger.WithField("cmd", "build-studio-info")

		k8sClient, _ := platformK8s.InitKubernetesClient()
		k8sRepoV2 := k8s.NewRepo(k8sClient, logger.WithField("context", "k8s-repo-v2"))

		customers := args

		if len(customers) > 0 && resetAll {
			logContext.Fatal("specify either a CUSTOMERID or '--all' flag")
		}

		if resetAll {
			logContext.Info("Discovering all customers from the platform")
			customers = extractCustomers(k8sRepoV2)
		}

		if len(customers) == 0 {
			logContext.Fatal("No customers found or no CUSTOMERID given")
		}

		filteredCustomer := filterCustomers(gitRepo, customers, platformEnvironment)
		logContext.Infof("Resetting studio configuration for customers: %v", filteredCustomer)

		studioConfig := storage.DefaultStudioConfig()
		if disabledEnvironments {
			studioConfig.DisabledEnvironments = []string{"*"}
		}

		if disableCanCreateApplication {
			studioConfig.CanCreateApplication = false
		}

		ResetStudioConfigs(gitRepo, filteredCustomer, studioConfig, logContext)
		logContext.Info("Done!")
	},
}

// ResetStudioConfigs resets all of the found customers studio.json files to enable automation for all environments
// and to enable overwriting
func ResetStudioConfigs(repo storage.Repo, customers []string, config platform.StudioConfig, logger logrus.FieldLogger) error {
	logContext := logger.WithFields(logrus.Fields{
		"function": "ResetStudioConfigs",
	})

	for _, customer := range customers {
		if err := repo.SaveStudioConfig(customer, config); err != nil {
			logContext.WithFields(logrus.Fields{
				"error":      err,
				"customerID": customer,
			}).Fatal("couldn't save and commit default studio config")
		}
	}
	return nil
}

func extractCustomers(repo k8s.Repo) []string {
	customers := make([]string, 0)
	namespaces, err := repo.GetNamespacesWithApplication()
	if err != nil {
		panic(err.Error())
	}

	for _, namespace := range namespaces {
		customerID := namespace.Annotations["dolittle.io/tenant-id"]
		customers = append(customers, customerID)
	}
	return customers
}

func filterCustomers(repo storage.Repo, customers []string, platformEnvironment string) []string {
	filtered := make([]string, 0)
	for _, customerID := range customers {
		customer, err := repo.GetTerraformTenant(customerID)
		if err != nil {
			continue
		}
		if customer.PlatformEnvironment != platformEnvironment {
			continue
		}
		filtered = append(filtered, customerID)
	}
	return filtered
}

func init() {
	studioCMD.Flags().Bool("disable-environments", false, "If flag set, Disable all environments")
	studioCMD.Flags().Bool("disable-create-application", false, "If flag set, Disable ability to create application")
	studioCMD.Flags().Bool("all", false, "Discovers all customers from the platform and resets all studio.json's to default state (everything allowed)")
}
