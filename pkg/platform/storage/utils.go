package storage

import (
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/thoas/go-funk"
)

// TODO rename once happy
func ConvertFromJSONApplication2(input JSONApplication2) platform.HttpResponseApplication {
	output := platform.HttpResponseApplication{
		ID:         input.ID,
		Name:       input.Name,
		TenantID:   input.TenantID,
		TenantName: input.TenantName,
	}

	for _, environment := range input.Environments {
		customerTenants := funk.Map(environment.Tenants, func(tenantID platform.TenantId) platform.TenantId {
			return platform.TenantId(tenantID)
		}).([]platform.TenantId)

		ingresses := make(map[platform.TenantId]platform.EnvironmentIngress)

		for _, customerTenantID := range environment.Tenants {
			mapped := make([]string, 0)
			for _, ingress := range environment.Ingresses {
				if funk.ContainsString(mapped, ingress.Host) {
					continue
				}

				mapped = append(mapped, ingress.Host)

				newIngress := platform.EnvironmentIngress{
					Host:         ingress.Host,
					DomainPrefix: ingress.DomainPrefix,
					SecretName:   ingress.SecretName,
				}

				ingresses[platform.TenantId(customerTenantID)] = newIngress
			}
		}

		newEnvironment := platform.HttpInputEnvironment{
			Name:          environment.Name,
			TenantID:      environment.TenantID,
			ApplicationID: environment.ApplicationID,
			Tenants:       customerTenants,
			Ingresses:     ingresses,
		}

		output.Environments = append(output.Environments, newEnvironment)
	}

	return output
}

func ConvertFromPlatformHttpResponseApplication(input platform.HttpResponseApplication) JSONApplication2 {
	output := JSONApplication2{
		ID:         input.ID,
		Name:       input.Name,
		TenantID:   input.TenantID,
		TenantName: input.TenantName,
	}

	for _, environment := range input.Environments {
		customerTenantIDS := funk.Map(environment.Tenants, func(tenantID platform.TenantId) string {
			return string(tenantID)
		}).([]string)

		customerTenantIngresses := make([]JSONEnvironmentIngress2, 0)

		for _, customerTenantID := range customerTenantIDS {
			ingresses := funk.Map(environment.Ingresses, func(envCustomerTenantID platform.TenantId, ingress platform.EnvironmentIngress) JSONEnvironmentIngress2 {
				return JSONEnvironmentIngress2{
					Host:        ingress.Host,
					Environment: environment.Name,
					//Path:             "TODO",
					CustomerTenantID: customerTenantID,
					DomainPrefix:     ingress.DomainPrefix,
					SecretName:       ingress.SecretName,
				}
			}).([]JSONEnvironmentIngress2)
			customerTenantIngresses = append(customerTenantIngresses, ingresses...)
		}

		// TODO build customer tenants
		customerTenants := make([]platform.CustomerTenantInfo, 0)
		newEnvironment := JSONEnvironment2{
			Name:            environment.Name,
			TenantID:        environment.TenantID,
			ApplicationID:   environment.ApplicationID,
			Tenants:         customerTenantIDS,
			Ingresses:       customerTenantIngresses,
			CustomerTenants: customerTenants,
		}

		output.Environments = append(output.Environments, newEnvironment)
	}
	return output
}
