package application

import (
	"fmt"

	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/dolittle/platform-api/pkg/platform/user"
	"github.com/spf13/cobra"
	"github.com/thoas/go-funk"
)

var accessCMD = &cobra.Command{
	Use:   "access",
	Short: "Access to the application",
	Long: `
	(Requires AZURE_XXX environment variables to be set)

	List users who have access to the application

	go run main.go tools application access {applicationID} list

	Add user who have access to the application

	go run main.go tools application access {applicationID} add human@dolittle.com
	`,
	Args: cobra.RangeArgs(2, 3),
	Run: func(cmd *cobra.Command, args []string) {
		settings, err := auth.GetSettingsFromEnvironment()
		if err != nil {
			fmt.Println(err.Error())
			fmt.Println("Missing AZURE_XXX settings")
			return
		}

		tenantID := settings.Values[auth.TenantID]
		authorizer := user.NewBearerAuthorizerFromEnvironmentVariables(settings)

		userClient := user.NewUserClient(tenantID, authorizer)
		groupClient := user.NewGroupsClient(tenantID, authorizer)
		activeDirectoryClient := user.NewUserActiveDirectory(groupClient, userClient)

		applicationID := args[0]

		do := args[1]
		if !funk.Contains([]string{"add", "remove", "list"}, do) {
			fmt.Println("Only allowed to add, remove or list")
			return
		}

		if do == "list" {
			getUsersByApplication(applicationID, activeDirectoryClient)
			return
		}

		if do == "add" {
			email := args[2]
			addUserToApplication(applicationID, email, activeDirectoryClient)
			return
		}

		if do == "remove" {
			email := args[2]
			removeUserFromApplication(applicationID, email, activeDirectoryClient)
			return
		}
	},
}

func getUsersByApplication(applicationID string, client user.UserActiveDirectory) {
	groupID, err := client.GetGroupIDByApplicationID(applicationID)
	if err != nil {
		fmt.Println("Failed to find application")
		return
	}

	users, err := client.GetUsersInApplication(groupID)

	if err != nil {
		fmt.Println("Failed to find users")
		return
	}

	for _, user := range users {
		fmt.Println(user.Email, user.ID)
	}
}

func addUserToApplication(applicationID string, email string, client user.UserActiveDirectory) {
	groupID, err := client.GetGroupIDByApplicationID(applicationID)
	if err != nil {
		fmt.Println(err)
		return
	}

	userID, err := client.GetUserIDByEmail(email)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = client.AddUserToGroup(userID, groupID)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func removeUserFromApplication(applicationID string, email string, client user.UserActiveDirectory) {
	groupID, err := client.GetGroupIDByApplicationID(applicationID)
	if err != nil {
		fmt.Println(err)
		return
	}

	userID, err := client.GetUserIDByEmail(email)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = client.RemoveUserFromGroup(userID, groupID)
	if err != nil {
		fmt.Println(err)
		return
	}
}
