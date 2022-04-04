package users

import (
	"fmt"

	"github.com/dolittle/platform-api/pkg/platform/user"
	"github.com/spf13/cobra"
)

var removeCMD = &cobra.Command{
	Use:   "remove",
	Short: "Remove a user from a customer in kratos",
	Long: `
	Remove a user with a customer by email:
	(Today this does not remove access to kubernetes, just Studio)

	go run main.go tools users remove --email="human@dolittle.com" --customer-id="fake-customer-id" --kratos-url="localhost:4434"
	`,
	Run: func(cmd *cobra.Command, args []string) {
		customerID, _ := cmd.Flags().GetString("customer-id")
		if customerID == "" {
			fmt.Println("A customerID is required")
			return
		}

		email, _ := cmd.Flags().GetString("email")

		if email == "" {
			fmt.Println("missing --email")
			return
		}

		err := kratosClient.RemoveCustomerToUserByEmail(email, customerID)
		if err != nil {
			// TODO this is duplicate, one should be removed
			if err == user.ErrNotFound {
				fmt.Println("Email could not be found")
				return
			}

			if err == user.ErrNotFound {
				fmt.Println("Customer and User are not connected")
				return
			}
			fmt.Println("error", err)
			return
		}

		fmt.Println("User can no longer login and access this customer via kratos")
	},
}

func init() {
	removeCMD.Flags().String("email", "", "Email address of the user")
	removeCMD.Flags().String("customer-id", "", "Customer Id to give the email access too")
}
