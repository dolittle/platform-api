package api

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var updateRepoCMD = &cobra.Command{
	Use:   "update-repo",
	Short: "Trigger pull on the git repo",
	Run: func(cmd *cobra.Command, args []string) {

		err := os.WriteFile("/tmp/trigger-git-pull", []byte(""), 0644)
		if err != nil {
			fmt.Println(err)
			return
		}

		//platformEnvironment := viper.GetString("tools.server.platformEnvironment")
		//gitRepoConfig := git.InitGit(logContext, platformEnvironment)
		//
		//os.Cre
		//gitRepo := gitStorage.NewGitStorage(
		//	logrus.WithField("context", "git-repo"),
		//	gitRepoConfig,
		//)
		//
		//err := gitRepo.Pull()
		//if err != nil {
		//	fmt.Println(err)
		//	return
		//}
	},
}
