package template

import (
	"github.com/spf13/cobra"
)

var RootCMD = &cobra.Command{
	Use:   "template",
	Short: "Create pre-filled Terraform HCL",
	Long:  ``,
}

func init() {
	RootCMD.AddCommand(applicationCMD)
	RootCMD.AddCommand(customerCMD)
}
