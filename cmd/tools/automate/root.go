package automate

import (
	"github.com/spf13/cobra"
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
	RootCmd.AddCommand(createApplicationCMD)
}
