package users

import (
	"fmt"

	"github.com/dolittle/platform-api/pkg/platform/user"
	"github.com/spf13/cobra"
)

var listEmailsCMD = &cobra.Command{
	Use:   "list-emails",
	Short: "List emails",
	Long: `
	Given an email get user

	go run main.go tools users list-emails
	`,
	Run: func(cmd *cobra.Command, args []string) {
		url := "http://localhost:4434/identities"
		// TODO look up users
		kratosUsers, err := user.GetUsersFromKratos(url)
		if err != nil {
			fmt.Println("Failed to get users")
			return
		}
		// Lookup email
		for _, kratosUser := range kratosUsers {
			fmt.Println(kratosUser.Traits.Email)
		}
	},
}
