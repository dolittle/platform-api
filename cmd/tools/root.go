package tools

import (
	"github.com/dolittle/platform-api/cmd/tools/application"
	"github.com/dolittle/platform-api/cmd/tools/automate"
	"github.com/dolittle/platform-api/cmd/tools/explore"
	"github.com/dolittle/platform-api/cmd/tools/job"
	"github.com/dolittle/platform-api/cmd/tools/m3connector"
	"github.com/dolittle/platform-api/cmd/tools/microservice"
	"github.com/dolittle/platform-api/cmd/tools/studio"
	"github.com/dolittle/platform-api/cmd/tools/terraform"
	"github.com/dolittle/platform-api/cmd/tools/users"
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
	RootCmd.AddCommand(explore.RootCmd)
	RootCmd.AddCommand(job.RootCMD)
	RootCmd.AddCommand(terraform.RootCMD)
	RootCmd.AddCommand(users.RootCMD)
	RootCmd.AddCommand(application.RootCMD)
	RootCmd.AddCommand(m3connector.RootCMD)
	RootCmd.AddCommand(microservice.RootCMD)
}
