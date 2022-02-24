package users

import (
	"fmt"

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
		kratosUsers, err := kratosClient.GetUsers()
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
