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
	Given an email add to a customerID

	go run main.go tools users add-user-to-customer --email="human@dolittle.com" --customer-id="fake-customer-id"
	`,
	Run: func(cmd *cobra.Command, args []string) {
		email, _ := cmd.Flags().GetString("email")
		if email == "" {
			fmt.Println("An email is required")
			return
		}

		customerID, _ := cmd.Flags().GetString("customer-id")
		if customerID == "" {
			fmt.Println("A customerID is required")
			return
		}

		outputCurl, _ := cmd.Flags().GetBool("output-curl")
		fmt.Println(outputCurl)

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

		exists := user.UserCustomersContains(found, customerID)

		if exists {
			fmt.Println("Email already has access to the customer id")
			return
		}

		// Add customer to Tenants
		found.Traits.Tenants = append(found.Traits.Tenants, customerID)
		// Print curl
		fmt.Println(user.PrintPutUser(url, found, customerID))
	},
}

func init() {
	addUserToCustomerCMD.Flags().String("email", "", "Email address of the user")
	addUserToCustomerCMD.Flags().String("customer-id", "", "Customer Id to give the email access too")
	addUserToCustomerCMD.PersistentFlags().Bool("ouput-curl", false, "Don't add the user, but instead output the curl command to do it")
}
