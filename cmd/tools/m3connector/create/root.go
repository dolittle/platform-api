package create

import (
	"github.com/spf13/cobra"
)

var RootCMD = &cobra.Command{
	Use:   "create",
	Short: "",
	Long:  ``,
}

func init() {
	RootCMD.AddCommand(environmentCMD)
}
