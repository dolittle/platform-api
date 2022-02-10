package cmd

import (
	"github.com/dolittle/platform-api/cmd/microservice"
	"github.com/dolittle/platform-api/cmd/rawdatalog"
	"github.com/dolittle/platform-api/cmd/tools"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "platform",
	Short: "Platform",
	Long:  `Entry point to different tools / services to help do things in the Dolittle platform`,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	viper.AutomaticEnv()
	rootCmd.AddCommand(microservice.RootCmd)
	rootCmd.AddCommand(rawdatalog.RootCmd)
	rootCmd.AddCommand(tools.RootCmd)

	viper.SetDefault("tools.jobs.image.operations", "dolittle/platform-operations:application-namespace")
	viper.SetDefault("tools.jobs.git.branch.local", "test-job")
	viper.SetDefault("tools.jobs.git.branch.remote", "test-job")
	viper.SetDefault("tools.jobs.git.user.name", "Auto Platform")
	viper.SetDefault("tools.jobs.git.user.email", "platform-auto@dolittle.com")

	viper.BindEnv("tools.jobs.image.operations", "JOBS_OPERATIONS_IMAGE")
	viper.BindEnv("tools.jobs.git.branch.local", "JOBS_GIT_BRANCH_LOCAL")
	viper.BindEnv("tools.jobs.git.branch.remote", "JOBS_GIT_BRANCH_LOCAL")
	viper.BindEnv("tools.jobs.git.user.name", "JOBS_GIT_USER_NAME")
	viper.BindEnv("tools.jobs.git.user.email", "JOBS_GIT_USER_EMAIL")
}
