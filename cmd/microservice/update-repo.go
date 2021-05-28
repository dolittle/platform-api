package microservice

import (
	"fmt"

	gitStorage "github.com/dolittle-entropy/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var updateRepoCMD = &cobra.Command{
	Use:   "update-repo",
	Short: "Trigger pull on the git repo",
	Run: func(cmd *cobra.Command, args []string) {
		gitRepoBranch := viper.GetString("tools.server.gitRepo.branch")
		if gitRepoBranch == "" {
			panic("GIT_BRANCH required")
		}
		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			"git@github.com:freshteapot/test-deploy-key.git",
			"/tmp/dolittle-k8s",
			gitRepoBranch,
			// TODO fix this, then update deployment
			"/Users/freshteapot/dolittle/.ssh/test-deploy",
		)

		err := gitRepo.Pull()
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}
