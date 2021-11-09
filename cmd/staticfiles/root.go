package staticfiles

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "static-files",
	Short: "Serve static files",
	Long:  ``,
}

func init() {
	RootCmd.AddCommand(serverCMD)
}
