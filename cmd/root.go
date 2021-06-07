package cmd

import (
	"github.com/dolittle-entropy/platform-api/cmd/microservice"
	"github.com/dolittle-entropy/platform-api/cmd/rawdatalog"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "tools",
	Short: "Platform tools",
	Long:  ``,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	viper.AutomaticEnv()
	rootCmd.AddCommand(microservice.RootCmd)
	rootCmd.AddCommand(rawdatalog.RootCmd)
}
