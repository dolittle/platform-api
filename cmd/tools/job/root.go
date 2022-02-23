package job

import (
	"github.com/dolittle/platform-api/cmd/tools/job/template"
	"github.com/spf13/cobra"
)

var RootCMD = &cobra.Command{
	Use:   "job",
	Short: "Commands to manage jobs",
	Long:  ``,
}

func init() {
	RootCMD.AddCommand(statusCMD)
	RootCMD.AddCommand(template.RootCMD)
}
