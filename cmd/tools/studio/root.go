package studio

import (
	"fmt"
	"log"
	"os"

	"github.com/dolittle/platform-api/cmd/tools/studio/cfg"
	"github.com/dolittle/platform-api/cmd/tools/studio/create"
	"github.com/dolittle/platform-api/cmd/tools/studio/get"
	"github.com/dolittle/platform-api/cmd/tools/studio/upsert"
	"github.com/dolittle/platform-api/pkg/git"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCmd = &cobra.Command{
	Use:   "studio",
	Short: "Studio specific tools",
	Long: `

Tools to help create files needed for studio from kubernetes and / or terraform`,
}

func init() {
	RootCmd.AddCommand(create.RootCMD)
	RootCmd.AddCommand(get.RootCMD)
	RootCmd.AddCommand(userAdminCMD)
	RootCmd.AddCommand(upsert.RootCMD)
	RootCmd.AddCommand(cfg.RootCMD)

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
