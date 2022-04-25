package application

import (
	"github.com/spf13/cobra"
)

var RootCMD = &cobra.Command{
	Use:   "application",
	Short: "Platform tools for application",
	Long: `

Tools to interact with applications in the platform`,
}

func init() {
	RootCMD.AddCommand(accessCMD)
}
