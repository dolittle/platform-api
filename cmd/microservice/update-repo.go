package microservice

import (
	"fmt"
	"os"

	gitStorage "github.com/dolittle-entropy/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var updateRepoCMD = &cobra.Command{
	Use:   "update-repo",
	Short: "Trigger pull on the git repo",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logContext := logrus.StandardLogger()
		gitRepoConfig := initGit(logContext)

		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoConfig,
			"/tmp/dolittle-k8s",
		)

		err := gitRepo.Pull()
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}
