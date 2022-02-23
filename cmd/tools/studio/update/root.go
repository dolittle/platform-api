package update

import (
	"github.com/spf13/cobra"
)

var RootCMD = &cobra.Command{
	Use:   "update",
	Short: "Update Studio resources",
	Long:  ``,
}

func init() {
	RootCMD.AddCommand(applicationCMD)
	RootCMD.AddCommand(studioCMD)
	RootCMD.AddCommand(terraformCMD)
}
