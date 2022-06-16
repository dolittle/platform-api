package create

import (
	"github.com/spf13/cobra"
)

var RootCMD = &cobra.Command{
	Use:   "create",
	Short: "Create Studio resources",
	Long:  ``,
}

func init() {
	RootCMD.AddCommand(serviceAccountCMD)
	RootCMD.AddCommand(microserviceCMD)
}
