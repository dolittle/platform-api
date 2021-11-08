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
}
