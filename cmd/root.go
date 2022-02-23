package cmd

import (
	"github.com/dolittle/platform-api/cmd/api"
	"github.com/dolittle/platform-api/cmd/rawdatalog"
	"github.com/dolittle/platform-api/cmd/template"
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
	rootCmd.AddCommand(api.RootCmd)
	rootCmd.AddCommand(rawdatalog.RootCmd)
	rootCmd.AddCommand(tools.RootCmd)
	rootCmd.AddCommand(template.RootCMD)

	viper.SetDefault("tools.server.platformEnvironment", "dev")
	viper.BindEnv("tools.server.platformEnvironment", "PLATFORM_ENVIRONMENT")
	rootCmd.PersistentFlags().String("platform-environment", viper.GetString("tools.server.platformEnvironment"), "Platform environment (dev or prod), not linked to application environment")
	viper.BindPFlag("tools.server.platformEnvironment", rootCmd.PersistentFlags().Lookup("platform-environment"))

	viper.SetDefault("tools.jobs.image.operations", "dolittle/platform-operations:latest")
	viper.SetDefault("tools.jobs.git.user.name", "Auto Platform")
	viper.SetDefault("tools.jobs.git.user.email", "platform-auto@dolittle.com")
	viper.SetDefault("tools.jobs.secrets.name", "dev-api-v1-secrets")
	viper.SetDefault("tools.jobs.git.remote.url", "git@github.com:dolittle-platform/Operations.git")
	viper.SetDefault("tools.jobs.git.remote.branch", "test-job")

	viper.BindEnv("tools.jobs.image.operations", "JOBS_OPERATIONS_IMAGE")
	viper.BindEnv("tools.jobs.git.user.name", "JOBS_GIT_USER_NAME")
	viper.BindEnv("tools.jobs.git.user.email", "JOBS_GIT_USER_EMAIL")
	viper.BindEnv("tools.jobs.secrets.name", "JOBS_SECRETS_NAME")
	viper.BindEnv("tools.jobs.git.remote.url", "JOBS_GIT_REMOTE_URL")
	viper.BindEnv("tools.jobs.git.remote.branch", "JOBS_GIT_REMOTE_BRANCH")
}
