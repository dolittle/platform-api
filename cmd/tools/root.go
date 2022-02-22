package tools

import (
	"github.com/dolittle/platform-api/cmd/tools/automate"
	"github.com/dolittle/platform-api/cmd/tools/explore"
	"github.com/dolittle/platform-api/cmd/tools/jobs"
	"github.com/dolittle/platform-api/cmd/tools/studio"
	"github.com/dolittle/platform-api/cmd/tools/terraform"
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
	RootCmd.AddCommand(automate.RootCmd)
	RootCmd.AddCommand(terraform.RootCmd)
	RootCmd.AddCommand(jobs.RootCmd)
	RootCmd.AddCommand(explore.RootCmd)
}
