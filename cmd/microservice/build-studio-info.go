package microservice

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	gitStorage "github.com/dolittle-entropy/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thoas/go-funk"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var buildStudioInfoCMD = &cobra.Command{
	Use:   "build-studio-info",
	Short: "Write studio info into the git repo",
	Long: `
	It will attempt to update the git repo with data from the cluster and skip those that have been setup.

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

		logContext.Info("Starting to extract applications from the cluster")
		applications := extractApplications(ctx, client)

		logContext.Info("Starting to reset all studio configs")
		ResetStudioConfigs(gitRepo, applications, shouldCommit, logContext)
		logContext.Info("Done!")
	},
}

// ResetStudioConfigs resets all of the found customers studio.json files to enable all automation for all environments
// and to enable overwriting
func ResetStudioConfigs(repo storage.Repo, applications []platform.HttpResponseApplication, shouldCommit bool, logger logrus.FieldLogger) error {
	logContext := logger.WithFields(logrus.Fields{
		"function": "ResetStudioConfigs",
	})
	logContext.Debug("Starting to reset all studio.json files to default values")

	for _, application := range applications {

		// filter out only this customers applications
		customersApplications := funk.Filter(applications, func(customerApplication platform.HttpResponseApplication) bool {
			return customerApplication.TenantID == application.TenantID
		}).([]platform.HttpResponseApplication)
		studioConfig := createDefaultStudioConfig(repo, application.TenantID, customersApplications)

		if shouldCommit {
			if err := repo.SaveStudioConfigAndCommit(application.TenantID, studioConfig); err != nil {
				logContext.WithFields(logrus.Fields{
					"error":    err,
					"tenantID": application.TenantID,
				}).Fatal("couldn't save and commit default studio config")
			}
		} else {
			if err := repo.SaveStudioConfig(application.TenantID, studioConfig); err != nil {
				logContext.WithFields(logrus.Fields{
					"error":    err,
					"tenantID": application.TenantID,
				}).Fatal("couldn't save default studio configs")
			}
		}
	}

	return nil
}

// createDefaultStudioConfig creates a studio.json file with default values
// set to enable automation and overwriting for that customer.
// The given applications will have all of their environments enabled for automation too.
func createDefaultStudioConfig(repo storage.Repo, customerID string, applications []platform.HttpResponseApplication) platform.StudioConfig {
	var environments []string
	for _, application := range applications {
		for _, environment := range application.Environments {
			applicationWithEnvironment := fmt.Sprintf("%s/%s", application.ID, strings.ToLower(environment.Name))
			environments = append(environments, applicationWithEnvironment)
		}
	}

	return platform.StudioConfig{
		BuildOverwrite:         true,
		AutomationEnabled:      true,
		AutomationEnvironments: environments,
	}
}

func init() {
	RootCmd.AddCommand(buildStudioInfoCMD)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	buildStudioInfoCMD.Flags().String("kube-config", fmt.Sprintf("%s/.kube/config", homeDir), "Full path to kubeconfig, set to incluster to make it use kubernetes lookup")
	viper.BindPFlag("tools.server.kubeConfig", buildStudioInfoCMD.Flags().Lookup("kube-config"))

	viper.BindEnv("tools.server.kubeConfig", "KUBECONFIG")

	buildStudioInfoCMD.Flags().Bool("commit", false, "Whether to commit and push the changes to the git repo")
	viper.BindPFlag("commit", buildStudioInfoCMD.Flags().Lookup("commit"))
}
