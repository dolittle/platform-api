package api

import (
	"fmt"
	"os"

	"github.com/dolittle/platform-api/pkg/git"
	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var updateRepoCMD = &cobra.Command{
	Use:   "update-repo",
	Short: "Trigger pull on the git repo",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logContext := logrus.StandardLogger()
		platformEnvironment := viper.GetString("tools.server.platformEnvironment")
		gitRepoConfig := git.InitGit(logContext, platformEnvironment)

		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		err := gitRepo.Pull()
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}
