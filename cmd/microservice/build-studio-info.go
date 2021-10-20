package microservice

import (
	"context"
	"os"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	gitStorage "github.com/dolittle-entropy/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var buildStudioInfoCMD = &cobra.Command{
	Use:   "build-studio-info [CUSTOMERID]... [FLAGS]",
	Short: "Resets the specified customers studio configuration",
	Long: `
	It will attempt to update the git repo with resetted studio configurations (studio.json).

	GIT_REPO_SSH_KEY="/Users/freshteapot/dolittle/.ssh/test-deploy" \
	GIT_REPO_BRANCH=auto-dev \
	GIT_REPO_URL="git@github.com:freshteapot/test-deploy-key.git" \
	go run main.go microservice build-studio-info --kube-config="/Users/freshteapot/.kube/config"
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

		shouldCommit := viper.GetBool("commit")
		resetAll := viper.GetBool("all")

		if len(args) == 0 {
			logContext.Fatal("no customerID given, did you mean to use '--all' flag?")
		}
		customers := args

		if len(customers) > 0 && resetAll {
			logContext.Fatal("specify either a CUSTOMERID or '--all' flag")
		}

		if resetAll {
			logContext.Info("Discovering all customers from the platform")
			customers = extractCustomers(ctx, client)
		}

		logContext.Infof("Resetting studio configuration for customers: %v", customers)
		ResetStudioConfigs(gitRepo, customers, shouldCommit, logContext)
		logContext.Info("Done!")
	},
}

// ResetStudioConfigs resets all of the found customers studio.json files to enable automation for all environments
// and to enable overwriting
func ResetStudioConfigs(repo storage.Repo, customers []string, shouldCommit bool, logger logrus.FieldLogger) error {
	logContext := logger.WithFields(logrus.Fields{
		"function": "ResetStudioConfigs",
	})

	defaultConfig := platform.StudioConfig{
		BuildOverwrite:       true,
		DisabledEnvironments: make([]string, 0),
	}
	for _, customer := range customers {
		if err := repo.SaveStudioConfig(customer, defaultConfig); err != nil {
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
	for _, namespace := range getNamespaces(ctx, client) {
		if isApplicationNamespace(namespace) {
			customerID := namespace.Annotations["dolittle.io/tenant-id"]
			customers = append(customers, customerID)
		}
	}
	return customers
}

func init() {
	RootCmd.AddCommand(buildStudioInfoCMD)

	buildStudioInfoCMD.Flags().Bool("all", false, "Discovers all customers from the platform and resets all studio.json's to default state (everything allowed)")
	viper.BindPFlag("all", buildStudioInfoCMD.Flags().Lookup("all"))
}
