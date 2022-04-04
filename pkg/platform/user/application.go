package user

import (
	"fmt"

	"github.com/dolittle/platform-api/pkg/platform/storage"
)

type UserRepository struct {
	kratos                KratosClientV5
	storageRepo           storage.Repo
	activeDirectoryClient UserActiveDirectory
}

func NewUserRepository() UserRepository {
	return UserRepository{}
}

// GetUsersInApplication return a list of users in the current application
func (r UserRepository) GetUsersInApplication(customerID string, applicationID string) ([]ActiveDirectoryUserInfo, error) {
	emptyInfo := make([]ActiveDirectoryUserInfo, 0)
	terraformApplication, err := r.storageRepo.GetTerraformApplication(customerID, applicationID)
	if err != nil {
		return emptyInfo, err
	}

	return r.activeDirectoryClient.GetUsersInApplication(terraformApplication.GroupID)
}

// TODO: Do we need kratos in this at all?
// AddUserToApplication
// Assumes email is in kratos.
// Assumes email is in activedirectory
func (r UserRepository) AddUserToApplication(customerID string, applicationID string, email string) error {
	terraformApplication, err := r.storageRepo.GetTerraformApplication(customerID, applicationID)
	if err != nil {
		return err
	}

	// Confirm the user is in kratos? (do we need to do this)
	_, err = r.kratos.GetUserByEmail(email)
	if err != nil {
		return err
	}

	// Lookup email in activedirectory now we know the user is in
	userID, err := r.activeDirectoryClient.GetUserIDByEmail(email)
	if err != nil {
		return err
	}

	fmt.Printf("Add user %s to active directory %s\n", userID, terraformApplication.GroupID)

	err = r.activeDirectoryClient.AddUserToGroup(userID, terraformApplication.GroupID)
	if err != nil {
		return err
	}

	return nil
}

func (r UserRepository) RemoveUserFromApplication(customerID string, applicationID string, email string) error {
	terraformApplication, err := r.storageRepo.GetTerraformApplication(customerID, applicationID)
	if err != nil {
		return err
	}

	// Lookup email in activedirectory now we know the user is in
	userID, err := r.activeDirectoryClient.GetUserIDByEmail(email)
	if err != nil {
		return err
	}

	fmt.Printf("Remove user %s to active directory %s\n", userID, terraformApplication.GroupID)

	err = r.activeDirectoryClient.RemoveUserFromGroup(userID, terraformApplication.GroupID)
	if err != nil {
		return err
	}

	return nil
}
