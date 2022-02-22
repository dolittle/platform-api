package helpers

import (
	"github.com/dolittle/platform-api/cmd/tools/helpers/commands"
	"github.com/dolittle/platform-api/cmd/tools/helpers/kubernetes"
	"github.com/dolittle/platform-api/cmd/tools/helpers/terraform"
	"github.com/spf13/cobra"
)

var RootCMD = &cobra.Command{
	Use:   "helpers",
	Short: "Helpers to create commands and HCL/YAML files to copy-paste",
	Long: `
Helpers to create commands and HCL/YAML files to copy-paste. These commands are meant to make our life easier by creating copy-paste ready commands.
	`,
}

func init() {
	RootCMD.AddCommand(terraform.RootCmd)
	RootCMD.AddCommand(kubernetes.RootCmd)
	RootCMD.AddCommand(commands.RootCmd)
}
