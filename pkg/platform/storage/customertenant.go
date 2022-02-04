package storage

import (
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/thoas/go-funk"
)

func GetCustomerTenants(application JSONApplication) []platform.CustomerTenantInfo {
	return funk.FlatMap(application.Environments, func(applicationEnvironment JSONEnvironment2) []platform.CustomerTenantInfo {
		return applicationEnvironment.CustomerTenants
	}).([]platform.CustomerTenantInfo)
}

func GetCustomerTenantsByEnvironment(application JSONApplication, environment string) []platform.CustomerTenantInfo {
	return funk.FlatMap(application.Environments, func(applicationEnvironment JSONEnvironment2) []platform.CustomerTenantInfo {
		if environment != applicationEnvironment.Name {
			return []platform.CustomerTenantInfo{}
		}
		return applicationEnvironment.CustomerTenants
	}).([]platform.CustomerTenantInfo)
}
