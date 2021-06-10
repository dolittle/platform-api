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

	viper.BindEnv("tools.server.gitRepo.branch", "GIT_BRANCH")
}
