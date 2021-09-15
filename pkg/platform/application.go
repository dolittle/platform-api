package platform

import (
	"fmt"

	"github.com/thoas/go-funk"
)

// NoConfiguredEnvironmentsError is the error that occurs when getting the tenant for an environment of a stored application
// that does not have any environments configured
type NoConfiguredEnvironmentsError struct {
	application HttpResponseApplication
}

func (e *NoConfiguredEnvironmentsError) Error() string {
	return fmt.Sprintf("There are no environments configured for application %s under customer %s", e.application.ID, e.application.TenantID)
}

// EnvironmentNotFoundError is the error that occurs when getting the tenant for an environment of a stored application
// that does not have the specific environment configured
type EnvironmentNotFoundError struct {
	application HttpResponseApplication
	environment string
}

func (e *EnvironmentNotFoundError) Error() string {
	return fmt.Sprintf("Environment %s is not configured for application %s under customer %s", e.environment, e.application.ID, e.application.TenantID)
}

// NoConfiguredTenantsError is the error that occurs when getting the tenant for an environment of a stored application
// that does not have any tenants configured
type NoConfiguredTenantsError struct {
	application HttpResponseApplication
	environment string
}

func (e *NoConfiguredTenantsError) Error() string {
	return fmt.Sprintf("Environment %s under application %s under customer %s does not have any configured tenants", e.environment, e.application.ID, e.application.TenantID)
}

// GetTenantForEnvironment gets the topmost tenant in the configured list of tenants for the application configuration
func (a HttpResponseApplication) GetTenantForEnvironment(environment string) (TenantId, error) {
	tenants, err := a.getTenantsInEnvironment(environment)
	if err != nil {
		return "", err
	}
	if len(tenants) == 0 {
		return "", &NoConfiguredTenantsError{a, environment}
	}
	return tenants[0], nil
}

func (a HttpResponseApplication) getTenantsInEnvironment(desiredEnvironmentName string) ([]TenantId, error) {
	environments := a.Environments
	if len(environments) == 0 {
		return make([]TenantId, 0), &NoConfiguredEnvironmentsError{a}
	}
	index := funk.IndexOf(environments, func(e HttpInputEnvironment) bool {
		return e.Name == desiredEnvironmentName
	})
	if index == -1 {
		return make([]TenantId, 0), &EnvironmentNotFoundError{a, desiredEnvironmentName}
	}
	return environments[index].Tenants, nil
}
