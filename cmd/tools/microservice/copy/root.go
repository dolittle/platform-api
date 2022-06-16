package copy

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCMD = &cobra.Command{
	Use:   "copy",
	Short: "Microservice config setup with Studio",
	Long:  ``,
}

func init() {
	RootCMD.AddCommand(environmentCMD)

	environmentCMD.PersistentFlags().String("application", "", "application id")
	environmentCMD.MarkPersistentFlagRequired("application")
	viper.BindPFlag("tools.studio.cfg.application", environmentCMD.PersistentFlags().Lookup("application"))

	environmentCMD.PersistentFlags().String("microservice-name", "", "Name of microservice to copy environment variables from")
	environmentCMD.MarkPersistentFlagRequired("microservice-name")
	viper.BindPFlag("tools.studio.cfg.microservice-name", environmentCMD.PersistentFlags().Lookup("microservice-name"))

	environmentCMD.PersistentFlags().String("from-env", "", "The environment to copy from")
	environmentCMD.MarkPersistentFlagRequired("from-env")
	viper.BindPFlag("tools.studio.cfg.from-env", environmentCMD.PersistentFlags().Lookup("from-env"))

	environmentCMD.PersistentFlags().String("to-env", "", "they environment to copy to")
	environmentCMD.MarkPersistentFlagRequired("to-env")
	viper.BindPFlag("tools.studio.cfg.to-env", environmentCMD.PersistentFlags().Lookup("to-env"))
}
