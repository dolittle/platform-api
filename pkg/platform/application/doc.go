package application

import (
	"github.com/dolittle/platform-api/pkg/platform"
)

type HttpResponseAccessUsers struct {
	ID    string                   `json:"id"`
	Name  string                   `json:"name"`
	Users []HttpResponseAccessUser `json:"users"`
}

type HttpResponseAccessUser struct {
	Email string `json:"email"`
}

type HttpInputAccessUser struct {
	Email string `json:"email"`
}

type HttpInputApplication struct {
	ID           string                            `json:"id"`
	Name         string                            `json:"name"`
	Environments []HttpInputApplicationEnvironment `json:"environments"`
}

type HttpInputApplicationEnvironmentCustomerTenant struct {
	ID string `json:"id"`
}

type HttpInputApplicationEnvironment struct {
	Name           string                                          `json:"name"`
	CustomerTenant []HttpInputApplicationEnvironmentCustomerTenant `json:"customerTenants"`
}

type HttpResponseApplication struct {
	ID            string                          `json:"id"`
	Name          string                          `json:"name"`
	CustomerID    string                          `json:"customerId"`
	CustomerName  string                          `json:"customerName"`
	Environments  []HttpResponseEnvironment       `json:"environments"`
	Microservices []platform.HttpMicroserviceBase `json:"microservices,omitempty"`
}

type HttpResponseEnvironment struct {
	AutomationEnabled bool   `json:"automationEnabled"`
	Name              string `json:"name"`
}

type HttpResponseApplications struct {
	// Customer ID
	ID string `json:"id"`
	// Customer Name
	Name                 string                              `json:"name"`
	CanCreateApplication bool                                `json:"canCreateApplication"`
	Applications         []platform.ShortInfoWithEnvironment `json:"applications"`
}
