package studio

import (
	"context"
	"fmt"
	"os"

	"github.com/dolittle/platform-api/pkg/git"
	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var buildApplicationInfoCMD = &cobra.Command{
	Use:   "build-application-info",
	Short: "Write application info into the git repo",
	Long: `
	It will attempt to update the git repo with data from the cluster and skip those that have been setup.

	GIT_REPO_BRANCH=dev \
	GIT_REPO_DRY_RUN=true \
	GIT_REPO_DIRECTORY="/tmp/dolittle-local-dev" \
	GIT_REPO_DIRECTORY_ONLY=true \
	go run main.go microservice build-application-info
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)
		logContext := logrus.StandardLogger()

		platformEnvironment, _ := cmd.Flags().GetString("platform-environment")
		gitRepoConfig := git.InitGit(logContext, platformEnvironment)

		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		ctx := context.TODO()
		client, _ := platformK8s.InitKubernetesClient()

		// TODO if the namespace had a label or annotation...
		// TODO Currently cheap to look up all
		logContext.Info("Starting to extract applications from the cluster")
		applications := extractApplications(ctx, client)

		filteredApplications := filterApplications(gitRepo, applications, platformEnvironment)

		logContext.Info(fmt.Sprintf("Saving %v application(s)", len(filteredApplications)))
		SaveApplications(gitRepo, filteredApplications, logContext)
		logContext.Info("Done!")
	},
}

// SaveApplications saves the Applications into applications.json and also creates a default studio.json if
// the customer doesn't have one
func SaveApplications(repo storage.Repo, applications []platform.HttpResponseApplication, logger logrus.FieldLogger) error {
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

func init() {
	buildApplicationInfoCMD.Flags().String("platform-environment", "dev", "Platform environment (dev or prod), not linked to application environment")
}
