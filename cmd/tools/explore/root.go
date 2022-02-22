package explore

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "explore",
	Short: "Explore the platform",
	Long: `

Tools to explore the platform`,
}

func init() {
	RootCmd.AddCommand(ingressCMD)
	RootCmd.AddCommand(dolittleResourcesCMD)
	RootCmd.AddCommand(microservicesCMD)
	RootCmd.AddCommand(jobStatusCMD)
	RootCmd.AddCommand(studioCustomersCMD)
}
