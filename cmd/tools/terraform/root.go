package terraform

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "terraform",
	Short: "Platform tools",
	Long: `

Tools to interact with the platform`,
}

func init() {
	RootCmd.AddCommand(createCustomerHclCMD)
	RootCmd.AddCommand(createCustomerApplicationHclCMD)
}
