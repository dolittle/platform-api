package m3connector

import (
	"github.com/dolittle/platform-api/cmd/tools/m3connector/create"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCMD = &cobra.Command{
	Use:   "m3connector",
	Short: "Platform tools for managing M3Connector and resources in Aiven tied to it",
	Long: `

Tools to interact with M3Connector and Aiven in the platform`,
}

func init() {
	RootCMD.AddCommand(create.RootCMD)

	viper.SetDefault("tools.m3connector.aiven.apiToken", "")
	viper.BindEnv("tools.m3connector.aiven.apiToken", "AIVEN_API_TOKEN")
	RootCMD.PersistentFlags().String("aiven-api-token", viper.GetString("tools.m3connector.aiven.apiToken"), "Aiven API token")
	viper.BindPFlag("tools.m3connector.aiven.apiToken", RootCMD.PersistentFlags().Lookup("aiven-api-token"))
}
