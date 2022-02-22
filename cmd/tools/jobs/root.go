package jobs

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Jobs to run in the platform",
}

func init() {
	RootCmd.AddCommand(createCustomerApplicationCMD)
	RootCmd.AddCommand(createCustomerCMD)
}
