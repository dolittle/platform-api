package delete

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "delete",
	Short: "Create pre-filled CLI commands for deleting resources from the platform",
	Long:  ``,
}

func init() {
	RootCmd.AddCommand(customerCMD)
	RootCmd.AddCommand(applicationCMD)

	RootCmd.PersistentFlags().String("directory", "", "Path to git repo")
}
