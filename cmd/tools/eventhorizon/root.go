package eventhorizon

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCMD = &cobra.Command{
	Use:   "eventhorizon",
	Short: "Tool for managing EventHorizon subscriptions",
	Long:  ``,
}

func init() {
	RootCMD.AddCommand(addCMD)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	RootCMD.PersistentFlags().String("kube-config", fmt.Sprintf("%s/.kube/config", homeDir), "Full path to kubeconfig, set to 'incluster' to make it use kubernetes lookup instead")
	viper.BindPFlag("tools.server.kubeConfig", RootCMD.PersistentFlags().Lookup("kube-config"))
	viper.BindEnv("tools.server.kubeConfig", "KUBECONFIG")
}
