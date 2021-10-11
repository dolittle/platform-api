package microservice

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCmd = &cobra.Command{
	Use:   "microservice",
	Short: "Microrservice tools",
	Long:  ``,
}

func init() {
	RootCmd.AddCommand(createCMD)
	RootCmd.AddCommand(buildTerraformInfoCMD)
	RootCmd.AddCommand(updateRepoCMD)
	RootCmd.AddCommand(gitTestCMD)

	viper.BindEnv("tools.server.gitRepo.sshKey", "GIT_REPO_SSH_KEY")
	viper.BindEnv("tools.server.gitRepo.branch", "GIT_REPO_BRANCH")
	viper.BindEnv("tools.server.gitRepo.url", "GIT_REPO_URL")
	viper.BindEnv("tools.server.gitRepo.directory", "GIT_REPO_DIRECTORY")
	viper.BindEnv("tools.server.gitRepo.directoryOnly", "GIT_REPO_DIRECTORY_ONLY")

	viper.SetDefault("tools.server.gitRepo.sshKey", "")
	viper.SetDefault("tools.server.gitRepo.url", "")
	viper.SetDefault("tools.server.gitRepo.directory", "/tmp/dolittle-k8s")
	viper.SetDefault("tools.server.gitRepo.directoryOnly", false)
}
