package users

import (
	"encoding/json"
	"fmt"

	"github.com/dolittle/platform-api/pkg/platform/user"
	"github.com/spf13/cobra"
)

var getUserCMD = &cobra.Command{
	Use:   "get-user",
	Short: "Get user by email",
	Long: `
	Given an email get user

	go run main.go tools users get-user --email="human@dolitte.com"
	`,
	Run: func(cmd *cobra.Command, args []string) {
		email, _ := cmd.Flags().GetString("email")
		if email == "" {
			fmt.Println("An email is required")
			return
		}

		url := "http://localhost:4434/identities"
		// TODO look up users
		kratosUsers, err := user.GetUsersFromKratos(url)
		if err != nil {
			fmt.Println("Failed to get users")
			return
		}
		// Lookup email

		found, err := user.GetUserFromListByEmail(kratosUsers, email)
		if err != nil {
			fmt.Println("Email not in system")
			return
		}
		b, _ := json.MarshalIndent(found, "", "  ")
		fmt.Println(string(b))
	},
}

func init() {
	getUserCMD.Flags().String("email", "", "Email address of the user")
}
