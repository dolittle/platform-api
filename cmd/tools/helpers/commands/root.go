package commands

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "commands",
	Short: "Create pre-filled CLI commands",
	Long:  ``,
}

func init() {
	RootCmd.AddCommand(deleteCustomerCMD)
	RootCmd.AddCommand(deleteApplicationCMD)
}
