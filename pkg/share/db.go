package share

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle"
)

type repo struct {
	all                            []dolittle.CustomerTenant
	lookupCustomerViaName          map[string]dolittle.CustomerTenant // name CustomerTenant
	lookupApplicationViaName       map[string]dolittle.Application    // applicationName
	lookupApplicationNameViaDomain map[string]string
}

func NewRepoFromJSON(raw []byte) *repo {
	r := &repo{
		all:                            []dolittle.CustomerTenant{},
		lookupCustomerViaName:          map[string]dolittle.CustomerTenant{},
		lookupApplicationViaName:       map[string]dolittle.Application{},
		lookupApplicationNameViaDomain: map[string]string{},
	}
	r.load(raw)
	return r
}

func (r *repo) load(raw []byte) {
	err := json.Unmarshal(raw, &r.all)
	if err != nil {
		panic(err)
	}

	for _, customer := range r.all {
		r.lookupCustomerViaName[customer.Tenant.Name] = customer
		r.lookupCustomerViaName[customer.Tenant.ID] = customer
	}

	for _, customer := range r.all {
		for _, application := range customer.Applications {
			key := fmt.Sprintf("%s/%s/%s", customer.Tenant.ID, application.Name, application.Environment)
			r.lookupApplicationViaName[key] = application

			for _, domain := range application.IngressHosts {
				r.lookupApplicationNameViaDomain[domain] = key
			}
		}
	}
}

func (r *repo) GetAll() []dolittle.CustomerTenant {
	return r.all
}

func (r *repo) FindApplication(tenantID string, application string, environment string) (dolittle.Application, error) {
	key := fmt.Sprintf("%s/%s/%s", tenantID, application, environment)
	app, ok := r.lookupApplicationViaName[key]
	if !ok {
		return dolittle.Application{}, errors.New("not-found")
	}
	return app, nil
}

func (r *repo) FindCustomerTenant(name string) (dolittle.CustomerTenant, error) {
	customer, ok := r.lookupCustomerViaName[name]
	if !ok {
		return dolittle.CustomerTenant{}, errors.New("not-found")
	}
	return customer, nil
}

func (r *repo) FindIngress(domain string) (dolittle.Application, error) {
	key, ok := r.lookupApplicationNameViaDomain[domain]
	if !ok {
		return dolittle.Application{}, errors.New("not-found")
	}

	// Duplicate with FindApplication
	app, ok := r.lookupApplicationViaName[key]
	if !ok {
		return dolittle.Application{}, errors.New("not-found")
	}
	return app, nil
}
