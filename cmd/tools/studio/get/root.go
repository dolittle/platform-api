package get

import (
	"github.com/spf13/cobra"
)

var RootCMD = &cobra.Command{
	Use:   "get",
	Short: "Get Studio resources",
	Long:  ``,
}

func init() {
	RootCMD.AddCommand(customersCMD)
}
