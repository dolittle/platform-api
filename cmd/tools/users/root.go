package users

import (
	"net/url"

	"github.com/dolittle/platform-api/pkg/platform/user"
	"github.com/spf13/cobra"
)

var (
	kratosClient user.KratosClientV5
)
var RootCMD = &cobra.Command{
	Use:   "users",
	Short: "Platform tools for users",
	Long: `

Tools to interact with the users in the platform`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		kratosURL, _ := cmd.Flags().GetString("kratos-url")

		kratosClient = user.NewKratosClientV5(&url.URL{
			Scheme: "http",
			Host:   kratosURL,
		})
	},
}

func init() {
	RootCMD.AddCommand(listEmailsCMD)
	RootCMD.AddCommand(getUserCMD)
	RootCMD.AddCommand(addCMD)
	RootCMD.AddCommand(removeCMD)

	RootCMD.PersistentFlags().String("kratos-url", "localhost:4434", "Url to kratos")
}
