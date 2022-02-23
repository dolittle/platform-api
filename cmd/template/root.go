package template

import (
	"github.com/dolittle/platform-api/cmd/template/delete"
	"github.com/spf13/cobra"
)

var RootCMD = &cobra.Command{
	Use:   "template",
	Short: "Create pre-filled templates for managing the platform and its resources",
	Long:  ``,
}

func init() {
	RootCMD.AddCommand(delete.RootCmd)
}
