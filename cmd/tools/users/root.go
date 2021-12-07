package users

import (
	"github.com/spf13/cobra"
)

var RootCMD = &cobra.Command{
	Use:   "users",
	Short: "Platform tools for users",
	Long: `

Tools to interact with the users in the platform`,
}

func init() {
	RootCMD.AddCommand(addUserToCustomerCMD)
	RootCMD.AddCommand(listEmailsCMD)
	RootCMD.AddCommand(getUserCMD)
}
