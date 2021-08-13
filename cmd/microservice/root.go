package microservice

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCmd = &cobra.Command{
	Use:   "microservice",
	Short: "Micorservice tools",
	Long:  ``,
}

func init() {
	RootCmd.AddCommand(createCMD)
	RootCmd.AddCommand(buildTerraformInfoCMD)
	RootCmd.AddCommand(updateRepoCMD)
	RootCmd.AddCommand(gitTestCMD)

	viper.BindEnv("tools.server.gitRepo.gitSshKey", "GIT_REPO_SSH_KEY")
	viper.BindEnv("tools.server.gitRepo.branch", "GIT_REPO_BRANCH")
	viper.BindEnv("tools.server.gitRepo.url", "GIT_REPO_URL")

	viper.SetDefault("tools.server.gitRepo.gitSshKey", "")
	viper.SetDefault("tools.server.gitRepo.url", "git@github.com:freshteapot/test-deploy-key.git")
}
