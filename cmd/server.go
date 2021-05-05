package cmd

import (
	"fmt"
	"os"

	"github.com/dolittle-entropy/platform-api/pkg/backup"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Server backed by customers.json database",
	Long: `
Server to static customer database

When we add:
- a new microservice
- a new tenant
- etc

We will want to rebuild the customers database and restart the server to reflect the
changes in the Studio. Very much replacable when we get a little more automation.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		secret := viper.GetString("tools.server.secret")
		pathToDB := viper.GetString("tools.server.pathToDb")

		if secret == "" {
			fmt.Println("secret is empty, assuming mistake")
			os.Exit(1)
		}

		if !utils.FileExists(pathToDB) {
			fmt.Println(fmt.Sprintf("path to db cant be found: %s", pathToDB))
			os.Exit(1)
		}

		backup.Run(secret, pathToDB)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	viper.BindEnv("tools.server.secret", "HEADER_SECRET")
	viper.BindEnv("tools.server.pathToDb", "PATH_TO_DB")
}
