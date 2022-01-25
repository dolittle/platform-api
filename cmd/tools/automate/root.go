package automate

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCmd = &cobra.Command{
	Use:   "automate",
	Short: "Commands to help semi-automation tasks in the cluster",
	Long: `

Tools to slowly help make it easier to live in a manual and semi-automated world`,
}

func init() {
	RootCmd.AddCommand(pullDolittleConfigCMD)
	RootCmd.AddCommand(getMicroservicesMetaDataCMD)
	RootCmd.AddCommand(addPlatformConfigCMD)
	RootCmd.AddCommand(importDolittleConfigMapsCMD)
	RootCmd.AddCommand(pullMicroserviceDeploymentCMD)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	RootCmd.PersistentFlags().String("kube-config", fmt.Sprintf("%s/.kube/config", homeDir), "Full path to kubeconfig, set to 'incluster' to make it use kubernetes lookup instead")
	viper.BindPFlag("tools.server.kubeConfig", RootCmd.PersistentFlags().Lookup("kube-config"))
	viper.BindEnv("tools.server.kubeConfig", "KUBECONFIG")
}
