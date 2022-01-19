package tools

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCmd = &cobra.Command{
	Use:   "tools",
	Short: "Platform tools",
	Long: `

Tools to interact with the platform`,
}

func init() {
	RootCmd.AddCommand(createCustomerHclCMD)
	RootCmd.AddCommand(pullDolittleConfigCMD)
	RootCmd.AddCommand(getMicroservicesMetaDataCMD)
	RootCmd.AddCommand(updateDolittleConfigCMD)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	RootCmd.PersistentFlags().String("kube-config", fmt.Sprintf("%s/.kube/config", homeDir), "Full path to kubeconfig, set to 'incluster' to make it use kubernetes lookup instead")
	viper.BindPFlag("tools.server.kubeConfig", RootCmd.PersistentFlags().Lookup("kube-config"))
	viper.BindEnv("tools.server.kubeConfig", "KUBECONFIG")
}
