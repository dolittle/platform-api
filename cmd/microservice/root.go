package microservice

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "microservice",
	Short: "Micorservice tools",
	Long:  ``,
}

func init() {
	RootCmd.AddCommand(createCMD)
}
