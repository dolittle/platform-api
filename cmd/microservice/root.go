package microservice

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCmd = &cobra.Command{
	Use:   "microservice",
	Short: "Microservice tools",
	Long:  ``,
}

func init() {
	RootCmd.AddCommand(createCMD)
	RootCmd.AddCommand(buildTerraformInfoCMD)
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

	viper.BindEnv("tools.server.gitRepo.sshKey", "GIT_REPO_SSH_KEY")
	viper.BindEnv("tools.server.gitRepo.branch", "GIT_REPO_BRANCH")
	viper.BindEnv("tools.server.gitRepo.url", "GIT_REPO_URL")
	viper.BindEnv("tools.server.gitRepo.directory", "GIT_REPO_DIRECTORY")
	viper.BindEnv("tools.server.gitRepo.directoryOnly", "GIT_REPO_DIRECTORY_ONLY")
	viper.BindEnv("tools.server.gitRepo.dryRun", "GIT_REPO_DRY_RUN")

	viper.SetDefault("tools.server.gitRepo.sshKey", "")
	viper.SetDefault("tools.server.gitRepo.url", "")
	viper.SetDefault("tools.server.gitRepo.directory", "/tmp/dolittle-k8s")
	viper.SetDefault("tools.server.gitRepo.directoryOnly", false)
	viper.SetDefault("tools.server.gitRepo.dryRun", false)
}
