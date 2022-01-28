package studio

import (
	"context"
	"os"

	"github.com/dolittle/platform-api/pkg/git"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/automate"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

var buildStudioInfoCMD = &cobra.Command{
	Use:   "build-studio-info [CUSTOMERID]... [FLAGS]",
	Short: "Resets the specified customers studio configuration",
	Long: `
	It will attempt to update the git repo with resetted studio configurations (studio.json).

	GIT_REPO_BRANCH=dev \
	GIT_REPO_DRY_RUN=true \
	GIT_REPO_DIRECTORY="/tmp/dolittle-local-dev" \
	GIT_REPO_DIRECTORY_ONLY=true \
	go run main.go tools studio build-studio-info
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logContext := logrus.StandardLogger()
		resetAll, _ := cmd.Flags().GetBool("all")
		disabledEnvironments, _ := cmd.Flags().GetBool("disable-environments")
		platformEnvironment, _ := cmd.Flags().GetString("platform-environment")
		gitRepoConfig := git.InitGit(logContext, platformEnvironment)

		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		ctx := context.TODO()
		k8sClient, _ := platformK8s.InitKubernetesClient()

		customers := args

		if len(customers) > 0 && resetAll {
			logContext.Fatal("specify either a CUSTOMERID or '--all' flag")
		}

		if resetAll {
			logContext.Info("Discovering all customers from the platform")
			customers = extractCustomers(ctx, k8sClient)
		}

		if len(customers) == 0 {
			logContext.Fatal("No customers found or no CUSTOMERID given")
		}

		filteredCustomer := filterCustomers(gitRepo, customers, platformEnvironment)
		logContext.Infof("Resetting studio configuration for customers: %v", filteredCustomer)

		studioConfig := GetConfig(disabledEnvironments)
		ResetStudioConfigs(gitRepo, filteredCustomer, studioConfig, logContext)
		logContext.Info("Done!")
	},
}

func GetConfig(disabledEnvironments bool) platform.StudioConfig {
	config := platform.StudioConfig{
		BuildOverwrite:       true,
		DisabledEnvironments: make([]string, 0),
	}

	if disabledEnvironments {
		config.DisabledEnvironments = append(config.DisabledEnvironments, "*")
	}
	return config
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

func extractCustomers(ctx context.Context, client kubernetes.Interface) []string {
	var customers []string
	for _, namespace := range automate.GetNamespaces(ctx, client) {
		if automate.IsApplicationNamespace(namespace) {
			customerID := namespace.Annotations["dolittle.io/tenant-id"]
			customers = append(customers, customerID)
		}
	}
	return customers
}

func init() {
	buildStudioInfoCMD.Flags().String("platform-environment", "dev", "Platform environment (dev or prod), not linked to application environment")
	buildStudioInfoCMD.Flags().Bool("disable-environments", false, "If flag set, Disable all environments")
	buildStudioInfoCMD.Flags().Bool("all", false, "Discovers all customers from the platform and resets all studio.json's to default state (everything allowed)")
}
