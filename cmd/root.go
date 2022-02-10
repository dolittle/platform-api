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
	// TODO move to yaml file could be nicer

	viper.SetDefault("tools.server.platformEnvironment", "dev")
	viper.BindEnv("tools.server.platformEnvironment", "PLATFORM_ENVIRONMENT")
	rootCmd.PersistentFlags().String("platform-environment", viper.GetString("tools.server.platformEnvironment"), "Platform environment (dev or prod), not linked to application environment")
	viper.BindPFlag("tools.server.platformEnvironment", rootCmd.PersistentFlags().Lookup("platform-environment"))

	viper.SetDefault("tools.jobs.image.operations", "dolittle/platform-operations:application-namespace")
	viper.SetDefault("tools.jobs.git.branch.local", "test-job")
	viper.SetDefault("tools.jobs.git.branch.remote", "test-job")
	viper.SetDefault("tools.jobs.git.user.name", "Auto Platform")
	viper.SetDefault("tools.jobs.git.user.email", "platform-auto@dolittle.com")
	viper.SetDefault("tools.jobs.secrets.name", "dev-api-v1-secrets") // TODO this is generic for the whole cluster, so doesn't care for the environment prefix

	viper.BindEnv("tools.jobs.image.operations", "JOBS_OPERATIONS_IMAGE")
	viper.BindEnv("tools.jobs.git.branch.local", "JOBS_GIT_BRANCH_LOCAL")
	viper.BindEnv("tools.jobs.git.branch.remote", "JOBS_GIT_BRANCH_LOCAL")
	viper.BindEnv("tools.jobs.git.user.name", "JOBS_GIT_USER_NAME")
	viper.BindEnv("tools.jobs.git.user.email", "JOBS_GIT_USER_EMAIL")
	viper.BindEnv("tools.jobs.secrets.name", "JOBS_SECRETS_NAME")
}
