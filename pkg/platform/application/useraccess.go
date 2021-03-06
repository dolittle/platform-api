package application

import (
	"github.com/dolittle/platform-api/pkg/platform/user"
)

type UserAccessRepo struct {
	azureActiveDirectory user.UserActiveDirectory
	kratos               user.KratosClientV5
}

func NewUserAccessRepo(kratos user.KratosClientV5, azureActiveDirectory user.UserActiveDirectory) UserAccessRepo {
	return UserAccessRepo{
		kratos:               kratos,
		azureActiveDirectory: azureActiveDirectory,
	}
}

func (r UserAccessRepo) GetUsers(applicationID string) ([]string, error) {
	users := make([]string, 0)

	groupID, err := r.azureActiveDirectory.GetGroupIDByApplicationID(applicationID)
	if err != nil {
		return users, err
	}

	currentUsers, err := r.azureActiveDirectory.GetUsersInApplication(groupID)

	if err != nil {
		return users, err
	}

	for _, currentUser := range currentUsers {
		users = append(users, currentUser.Email)
	}

	return users, nil
}

func (r UserAccessRepo) AddUser(customerID string, applicationID string, email string) error {
	// Add to kratos, fail silently if already there

	err := r.kratos.AddCustomerToUserByEmail(email, customerID)
	if err != nil {
		if err != user.ErrCustomerUserConnectionAlreadyExists {
			return err
		}
	}

	// Add to azure
	groupID, err := r.azureActiveDirectory.GetGroupIDByApplicationID(applicationID)
	if err != nil {
		return err
	}

	// TODO this might need more logic to handle when the email is already there
	return r.azureActiveDirectory.AddUserToGroupByEmail(email, groupID)
}

func (r UserAccessRepo) RemoveUser(applicationID string, email string) error {
	// Remove from azure
	groupID, err := r.azureActiveDirectory.GetGroupIDByApplicationID(applicationID)
	if err != nil {
		return err
	}

	return r.azureActiveDirectory.RemoveUserFromGroupByEmail(email, groupID)
}
