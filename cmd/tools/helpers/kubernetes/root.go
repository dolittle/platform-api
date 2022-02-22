package kubernetes

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "kubernetes",
	Short: "Create pre-filled Kubernetes YAML files",
	Long:  ``,
}

func init() {
	RootCmd.AddCommand(applicationJobCMD)
	RootCmd.AddCommand(customerJobCMD)
}
