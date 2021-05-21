package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/dolittle-entropy/platform-api/pkg/backup"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rebuildCmd = &cobra.Command{
	Use:   "rebuild-db",
	Short: "Rebuild the database for the server",
	Long: `Build the database based on what is in the cluster today
go run main.go rebuild-db --kube-confg="" --with-secrets | jq > customers.json
	`,
	Run: func(cmd *cobra.Command, args []string) {
		kubeconfig := viper.GetString("tools.db.kubeConfig")
		withSecrets := viper.GetBool("tools.db.withSecrets")

		data := backup.Rebuild(kubeconfig, withSecrets)
		b, _ := json.Marshal(data)
		fmt.Println(string(b))
	},
}

func init() {
	rootCmd.AddCommand(rebuildCmd)
	rebuildCmd.Flags().String("kube-config", "", "FullPath to kubeconfig")
	rebuildCmd.Flags().Bool("with-secrets", false, "Include Sensitive information")
	viper.BindPFlag("tools.db.kubeConfig", rebuildCmd.Flags().Lookup("kube-config"))
	viper.BindPFlag("tools.db.withSecrets", rebuildCmd.Flags().Lookup("with-secrets"))
}
