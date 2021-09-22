package tool

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "tool",
	Short: "A collection of tools to help do things in the platform",
}

func init() {
	RootCmd.AddCommand(deleteMicroserviceCMD)
}
