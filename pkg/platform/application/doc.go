package application

import (
	"github.com/dolittle/platform-api/pkg/platform"
)

type HttpInputApplication struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Environments []string `json:"environments"`
}

type HttpResponseApplication struct {
	ID            string                          `json:"id"`
	Name          string                          `json:"name"`
	TenantID      string                          `json:"tenantId"`   // TODO why do we include this?
	TenantName    string                          `json:"tenantName"` // TODO why do we include this?
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
