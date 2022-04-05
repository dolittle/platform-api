package user

import "errors"

type KratosUser struct {
	ID        string           `json:"id"`
	SchemaID  string           `json:"schema_id"`
	SchemaURL string           `json:"schema_url"`
	Traits    KratosUserTraits `json:"traits"`
}

type KratosUserTraits struct {
	Email   string   `json:"email"`
	Tenants []string `json:"tenants"`
}

var (
	ErrCustomerUserConnectionAlreadyExists = errors.New("customer-user-connection-already-exists")
	ErrNotFound                            = errors.New("not-found")
	ErrTooManyResults                      = errors.New("too-many-results")
	ErrEmailAlreadyExists                  = errors.New("application-email-already-exists")
)
