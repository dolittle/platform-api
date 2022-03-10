package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// SetupStringConfiguration makes it easier to setup the flags, keys, env variables and defaults for viper & cobra
func SetupStringConfiguration(cmd *cobra.Command, key, flag, envVarName, defaultValue, description string) {
	viper.SetDefault(key, defaultValue)
	viper.BindEnv(key, envVarName)
	cmd.PersistentFlags().String(flag, defaultValue, description)
	viper.BindPFlag(key, cmd.PersistentFlags().Lookup(flag))
}
