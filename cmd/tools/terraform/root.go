package terraform

import (
	"github.com/dolittle/platform-api/cmd/tools/terraform/template"
	"github.com/spf13/cobra"
)

var RootCMD = &cobra.Command{
	Use:   "terraform",
	Short: "Manage Terraform",
	Long:  ``,
}

func init() {
	RootCMD.AddCommand(template.RootCMD)
}
