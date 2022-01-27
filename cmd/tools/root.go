package tools

import (
	"github.com/dolittle/platform-api/cmd/tools/automate"
	"github.com/dolittle/platform-api/cmd/tools/studio"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "tools",
	Short: "Platform tools",
	Long: `

Tools to interact with the platform`,
}

func init() {
	RootCmd.AddCommand(studio.RootCmd)
	RootCmd.AddCommand(createCustomerHclCMD)
	RootCmd.AddCommand(automate.RootCmd)
}
