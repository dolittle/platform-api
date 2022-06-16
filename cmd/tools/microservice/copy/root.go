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

	environmentCMD.Flags().String("application", "", "application id")
	environmentCMD.MarkPersistentFlagRequired("application")
	viper.BindPFlag("tools.studio.cfg.application", environmentCMD.Flags().Lookup("application"))

	environmentCMD.Flags().String("microservice-name", "", "Name of microservice to copy environment variables from")
	environmentCMD.MarkPersistentFlagRequired("microservice-name")
	viper.BindPFlag("tools.studio.cfg.microservice-name", environmentCMD.Flags().Lookup("microservice-name"))

	environmentCMD.Flags().String("from-env", "", "The environment to copy from")
	environmentCMD.MarkPersistentFlagRequired("from-env")
	viper.BindPFlag("tools.studio.cfg.from-env", environmentCMD.Flags().Lookup("from-env"))

	environmentCMD.Flags().String("to-env", "", "they environment to copy to")
	environmentCMD.MarkPersistentFlagRequired("to-env")
	viper.BindPFlag("tools.studio.cfg.to-env", environmentCMD.Flags().Lookup("to-env"))
}
