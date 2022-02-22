package explore

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "explore",
	Short: "Explore the Cluster",
	Long: `

Tools to Explore the cluster`,
}

func init() {
	RootCmd.AddCommand(ingressCMD)
	RootCmd.AddCommand(dolittleResourcesCMD)
	RootCmd.AddCommand(microservicesCMD)
	RootCmd.AddCommand(jobStatusCMD)
}
