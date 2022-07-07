package m3connector

import (
	"fmt"
	"log"
	"os"

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

	viper.SetDefault("tools.m3connector.aiven.project", "")
	viper.BindEnv("tools.m3connector.aiven.project", "AIVEN_PROJECT")
	RootCMD.PersistentFlags().String("aiven-project", viper.GetString("tools.m3connector.aiven.project"), "Aiven project")
	viper.BindPFlag("tools.m3connector.aiven.project", RootCMD.PersistentFlags().Lookup("aiven-project"))

	viper.SetDefault("tools.m3connector.aiven.service", "")
	viper.BindEnv("tools.m3connector.aiven.service", "AIVEN_SERVICE")
	RootCMD.PersistentFlags().String("aiven-service", viper.GetString("tools.m3connector.aiven.service"), "Aiven service")
	viper.BindPFlag("tools.m3connector.aiven.service", RootCMD.PersistentFlags().Lookup("aiven-service"))

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	RootCMD.PersistentFlags().String("kube-config", fmt.Sprintf("%s/.kube/config", homeDir), "Full path to kubeconfig, set to 'incluster' to make it use kubernetes lookup instead")
	viper.BindPFlag("tools.server.kubeConfig", RootCMD.PersistentFlags().Lookup("kube-config"))
	viper.BindEnv("tools.server.kubeConfig", "KUBECONFIG")
}
