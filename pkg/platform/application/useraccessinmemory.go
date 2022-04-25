package application

import "fmt"

type UserAccessRepoInMemory struct {
	store map[string]string
}

func NewUserAccessRepoInMemory() UserAccessRepoInMemory {
	return UserAccessRepoInMemory{
		store: make(map[string]string),
	}
}

func (r UserAccessRepoInMemory) GetUsers(applicationID string) ([]string, error) {
	users := make([]string, 0)
	for _, v := range r.store {
		users = append(users, v)
	}

	return users, nil
}

func (r UserAccessRepoInMemory) AddUser(customerID string, applicationID string, email string) error {
	key := fmt.Sprintf("%s:%s", applicationID, email)
	r.store[key] = email
	return nil
}

func (r UserAccessRepoInMemory) RemoveUser(applicationID string, email string) error {
	key := fmt.Sprintf("%s:%s", applicationID, email)
	delete(r.store, key)
	return nil
}
