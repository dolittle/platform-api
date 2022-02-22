package terraform

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "terraform",
	Short: "Create pre-filled Terraform HCL files",
	Long:  ``,
}

func init() {
	RootCmd.AddCommand(applicationCMD)
	RootCmd.AddCommand(customerCMD)
}
