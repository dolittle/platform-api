package tools

import (
	"fmt"
	"os"

	"github.com/dolittle/platform-api/pkg/git"
	gitStorage "github.com/dolittle/platform-api/pkg/platform/storage/git"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var getHeadImageCMD = &cobra.Command{
	Use:   "get-head-image",
	Short: "Write terraform output for a customer",
	Long: `
	Outputs a new Dolittle platform customer in hcl to stdout.

	go run main.go tools get-head-image
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// Lookup image based on microserviceID?
		// Lookup image based on application/env/microserviceID?
		// Get all
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logContext := logrus.StandardLogger()
		gitRepoConfig := git.InitGit(logContext)

		gitRepo := gitStorage.NewGitStorage(
			logrus.WithField("context", "git-repo"),
			gitRepoConfig,
		)

		fmt.Println(gitRepo)
	},
}

func init() {
	git.SetupViper()
}
