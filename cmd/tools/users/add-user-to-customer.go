package users

import (
	"fmt"

	"github.com/dolittle/platform-api/pkg/platform/user"
	"github.com/spf13/cobra"
)

var addUserToCustomerCMD = &cobra.Command{
	Use:   "add-user-to-customer",
	Short: "Add user to customer",
	Long: `
	Connnect a user with a customer by email:

	go run main.go tools users add-user-to-customer --email="human@dolittle.com" --customer-id="fake-customer-id"


	Connnect a user with a customer by user-id:

	go run main.go tools users add-user-to-customer --user-id="fake-user-id" --customer-id="fake-customer-id"
	`,
	Run: func(cmd *cobra.Command, args []string) {
		customerID, _ := cmd.Flags().GetString("customer-id")
		if customerID == "" {
			fmt.Println("A customerID is required")
			return
		}

		userID, _ := cmd.Flags().GetString("user-id")
		email, _ := cmd.Flags().GetString("email")

		if userID == "" && email == "" {
			fmt.Println("missing --email or --user-id")
			return
		}

		if userID != "" && email != "" {
			fmt.Println("only --email or --user-id is allowed, not both")
			return
		}

		if userID != "" {
			err := kratosClient.AddCustomerToUserByUserID(userID, customerID)
			if err != nil {
				if err == user.ErrNotFound {
					fmt.Println("User could not be found")
					return
				}

				if err == user.ErrCustomerUserConnectionAlreadyExists {
					fmt.Println("Customer and User already connected")
					return
				}
				fmt.Println("error", err)
				return
			}

			fmt.Println("User can now login and access this customer via kratos")
			return
		}

		if email != "" {
			err := kratosClient.AddCustomerToUserByEmail(email, customerID)
			if err != nil {
				if err == user.ErrNotFound {
					fmt.Println("Email could not be found")
					return
				}

				if err == user.ErrCustomerUserConnectionAlreadyExists {
					fmt.Println("Customer and User already connected")
					return
				}
				fmt.Println("error", err)
				return
			}

			fmt.Println("User can now login and access this customer via kratos")
			return
		}
	},
}

func init() {
	addUserToCustomerCMD.Flags().String("email", "", "Email address of the user")
	addUserToCustomerCMD.Flags().String("user-id", "", "id of user")
	addUserToCustomerCMD.Flags().String("customer-id", "", "Customer Id to give the email access too")
}
