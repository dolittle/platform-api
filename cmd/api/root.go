package api

import (
	"fmt"
	"log"
	"os"

	"github.com/dolittle/platform-api/pkg/git"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCmd = &cobra.Command{
	Use:   "api",
	Short: "api for controlling the platform",
	Long:  ``,
}

func init() {
	RootCmd.AddCommand(updateRepoCMD)
	RootCmd.AddCommand(gitTestCMD)

	RootCmd.PersistentFlags().Bool("git-dry-run", false, "Don't commit and push changes")
	viper.BindPFlag("tools.server.gitRepo.dryRun", RootCmd.PersistentFlags().Lookup("git-dry-run"))

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	RootCmd.PersistentFlags().String("kube-config", fmt.Sprintf("%s/.kube/config", homeDir), "Full path to kubeconfig, set to 'incluster' to make it use kubernetes lookup instead")
	viper.BindPFlag("tools.server.kubeConfig", RootCmd.PersistentFlags().Lookup("kube-config"))
	viper.BindEnv("tools.server.kubeConfig", "KUBECONFIG")

	git.SetupViper()
}
