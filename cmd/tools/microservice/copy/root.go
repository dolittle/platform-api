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
	RootCMD.AddCommand(heyCMD)

	heyCMD.PersistentFlags().String("application", "", "application id")
	heyCMD.MarkPersistentFlagRequired("application")
	viper.BindPFlag("tools.studio.cfg.application", heyCMD.PersistentFlags().Lookup("application"))

	heyCMD.PersistentFlags().String("microservice-name", "", "Name of microservice to copy environment variables from")
	heyCMD.MarkPersistentFlagRequired("microservice-name")
	viper.BindPFlag("tools.studio.cfg.microservice-name", heyCMD.PersistentFlags().Lookup("microservice-name"))

	heyCMD.PersistentFlags().String("from-env", "", "The environment to copy from")
	heyCMD.MarkPersistentFlagRequired("from-env")
	viper.BindPFlag("tools.studio.cfg.from-env", heyCMD.PersistentFlags().Lookup("from-env"))

	heyCMD.PersistentFlags().String("to-env", "", "they environment to copy to")
	heyCMD.MarkPersistentFlagRequired("to-env")
	viper.BindPFlag("tools.studio.cfg.to-env", heyCMD.PersistentFlags().Lookup("to-env"))
}
