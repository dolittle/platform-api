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
		kratosClient = user.NewKratosClientV5(&url.URL{
			Scheme: "http",
			Host:   "localhost:4434",
		})
	},
}

func init() {
	RootCMD.AddCommand(addUserToCustomerCMD)
	RootCMD.AddCommand(listEmailsCMD)
	RootCMD.AddCommand(getUserCMD)
}
