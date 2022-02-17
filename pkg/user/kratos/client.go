package kratos

import (
	"errors"
	"net/url"

	ory "github.com/ory/kratos-client-go"
)

type Client interface {
	GetUser(userUUID string) error
}

type client struct {
	client *ory.OryKratos
}

func NewClient(publicEndpoint *url.URL) (Client, error) {
	config := ory.DefaultTransportConfig().WithSchemes([]string{publicEndpoint.Scheme}).WithHost(publicEndpoint.Host).WithBasePath(publicEndpoint.Path)
	ory.
	client := &client{
		client: ory.NewHTTPClientWithConfig(nil, config),
	}
	return client, nil
}

func (c *client) GetUser(userUUID string) error {
	return errors.New("TODO")
}
