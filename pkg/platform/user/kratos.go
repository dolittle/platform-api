package user

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	kratosClient "github.com/ory/kratos-client-go/client"
	"github.com/ory/kratos-client-go/client/admin"
	"github.com/ory/kratos-client-go/models"

	"github.com/thoas/go-funk"
)

type KratosClientV5 interface {
	GetUsers() ([]KratosUser, error)
	GetUser(id string) (KratosUser, error)
	UpdateUser(user KratosUser) error
	AddCustomerToUser(user KratosUser, customerID string) error
	AddCustomerToUserByUserID(userID string, customerID string) error
	AddCustomerToUserByEmail(email string, customerID string) error
}

var (
	ErrCustomerUserConnectionAlreadyExists = errors.New("customer-user-connection-already-exists")
	ErrNotFound                            = errors.New("not-found")
)

type kratosClientV5 struct {
	client *kratosClient.OryKratos
}

func NewKratosClientV5(clientURL *url.URL) KratosClientV5 {
	config := kratosClient.DefaultTransportConfig().WithSchemes([]string{clientURL.Scheme}).WithHost(clientURL.Host).WithBasePath(clientURL.Path)
	return kratosClientV5{
		client: kratosClient.NewHTTPClientWithConfig(nil, config),
	}
}

func (c kratosClientV5) AddCustomerToUserByUserID(userID string, customerID string) error {
	kratosUser, err := c.GetUser(userID)
	if err != nil {
		return err
	}

	return c.AddCustomerToUser(kratosUser, customerID)
}

func (c kratosClientV5) AddCustomerToUserByEmail(email string, customerID string) error {
	kratosUsers, err := c.GetUsers()
	if err != nil {
		return err
	}

	kratosUser, err := GetUserFromListByEmail(kratosUsers, email)
	if err != nil {
		return err
	}

	return c.AddCustomerToUser(kratosUser, customerID)
}

func (c kratosClientV5) AddCustomerToUser(kratosUser KratosUser, customerID string) error {
	exists := UserCustomersContains(kratosUser, customerID)

	if exists {
		return ErrCustomerUserConnectionAlreadyExists
	}

	kratosUser.Traits.Tenants = append(kratosUser.Traits.Tenants, customerID)

	err := c.UpdateUser(kratosUser)
	if err != nil {
		return err
	}
	return nil
}

func (c kratosClientV5) UpdateUser(user KratosUser) error {
	_, err := c.client.Admin.UpdateIdentity(&admin.UpdateIdentityParams{
		ID: user.ID,
		Body: &models.UpdateIdentity{
			Traits: user.Traits,
		},
		Context: context.TODO(),
	})
	return err
}

func (c kratosClientV5) convertKratosUserModelPayloadToKratosUser(model *models.Identity) (KratosUser, error) {
	var traits KratosUserTraits
	kratosUser := KratosUser{}

	traitsData := model.Traits
	b, err := json.Marshal(traitsData)
	if err != nil {
		return kratosUser, err
	}

	err = json.Unmarshal(b, &traits)
	if err != nil {
		return kratosUser, err
	}

	kratosUser = KratosUser{
		ID:        string(model.ID),
		SchemaID:  *model.SchemaID,
		SchemaURL: *model.SchemaURL,
		Traits:    traits,
	}
	return kratosUser, err
}

func (c kratosClientV5) GetUser(id string) (KratosUser, error) {
	i, err := c.client.Admin.GetIdentity(admin.NewGetIdentityParams().WithID(id))
	if err != nil {
		// This api is horrible :P, guess thats why it has been updated
		if strings.Contains(err.Error(), "404") {
			return KratosUser{}, ErrNotFound
		}

		return KratosUser{}, err
	}

	return c.convertKratosUserModelPayloadToKratosUser(i.Payload)
}

func (c kratosClientV5) GetUsers() ([]KratosUser, error) {
	kratosUsers := make([]KratosUser, 0)
	// Possible issue with pagination, not clear
	items, err := c.client.Admin.ListIdentities(nil)

	if err != nil {
		return kratosUsers, err
	}

	for _, item := range items.Payload {
		kratosUser, err := c.convertKratosUserModelPayloadToKratosUser(item)
		if err != nil {
			continue
		}
		kratosUsers = append(kratosUsers, kratosUser)
	}
	return kratosUsers, nil
}

func GetUserFromListByEmail(users []KratosUser, email string) (KratosUser, error) {
	found := funk.Find(users, func(kratosUser KratosUser) bool {
		return kratosUser.Traits.Email == email
	})

	if found == nil {
		return KratosUser{}, ErrNotFound
	}

	return found.(KratosUser), nil
}

func UserCustomersContains(user KratosUser, customerID string) bool {
	exists := funk.Contains(user.Traits.Tenants, func(currentCustomerID string) bool {
		return customerID == currentCustomerID
	})

	return exists
}
