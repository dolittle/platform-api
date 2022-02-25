package upsert

import (
	"github.com/spf13/cobra"
)

var RootCMD = &cobra.Command{
	Use:   "upsert",
	Short: "Upsert Studio resources",
	Long:  ``,
}

func init() {
	RootCMD.AddCommand(applicationCMD)
	RootCMD.AddCommand(terraformCMD)
	RootCMD.AddCommand(studioCMD)
}
