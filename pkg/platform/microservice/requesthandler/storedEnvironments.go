package requesthandler

import (
	_ "embed"
	"fmt"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"github.com/thoas/go-funk"
)

type storedEnvironments struct {
	gitRepo storage.Repo
}

func NewStoredEnvironments(gitRepo storage.Repo) StoredEnvironments {
	return &storedEnvironments{gitRepo}
}

func (e *storedEnvironments) GetTenant(customerID, applicationID, environment string) (platform.TenantId, error) {
	application, err := e.gitRepo.GetApplication(customerID, applicationID)
	if err != nil {
		return "", err
	}

	tenant, err := application.GetTenantForEnvironment(environment)
	if err != nil {
		return "", err
	}
	return tenant, nil
}

func (e *storedEnvironments) GetIngress(customerID, applicationID, environment string) (platform.EnvironmentIngress, error) {
	ingress := platform.EnvironmentIngress{}
	application, err := e.gitRepo.GetApplication(customerID, applicationID)
	if err != nil {
		return ingress, err
	}
	tenant, err := application.GetTenantForEnvironment(environment)
	if err != nil {
		return ingress, err
	}
	ingress, ok := application.Environments[funk.IndexOf(application.Environments, func(e platform.HttpInputEnvironment) bool {
		return e.Name == environment
	})].Ingresses[tenant]
	if !ok {
		return ingress, fmt.Errorf("Failed to get stored ingress for tenant %s in environment %s", string(tenant), environment)
	}
	return ingress, nil
}
