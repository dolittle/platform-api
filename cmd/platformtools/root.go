package platformtools

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "platform-tools",
	Short: "tools",
	Long:  ``,
}

func init() {
	RootCmd.AddCommand(stubCustomerCMD)

}
