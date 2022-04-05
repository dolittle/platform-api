package users

import (
	"encoding/json"
	"fmt"

	"github.com/dolittle/platform-api/pkg/platform/user"

	"github.com/spf13/cobra"
)

var getUserCMD = &cobra.Command{
	Use:   "get-user",
	Short: "Get user by email in kratos",
	Long: `
	Given an email get user

	go run main.go tools users get-user --email="human@dolittle.com" --kratos-url="localhost:4434"
	`,
	Run: func(cmd *cobra.Command, args []string) {

		email, _ := cmd.Flags().GetString("email")
		if email == "" {
			fmt.Println("An email is required")
			return
		}

		kratosUsers, err := kratosClient.GetUsers()
		if err != nil {
			fmt.Println("Failed to get users")
			return
		}

		kratosUser, err := user.GetUserFromListByEmail(kratosUsers, email)
		if err != nil {
			if err == user.ErrNotFound {
				fmt.Println("Email could not be found")
				return
			}
			fmt.Println("error", err)
			return
		}

		b, _ := json.MarshalIndent(kratosUser, "", "  ")
		fmt.Println(string(b))
	},
}

func init() {
	getUserCMD.Flags().String("email", "", "Email address of the user")
}
