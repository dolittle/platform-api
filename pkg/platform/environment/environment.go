package environment

func AddTenant(applicationID, environment, tenantID, subdomain string) {
	namespace := fmt.

}

func AddConsumerTenant(
	applicationID,
	microserviceID,
	environment,
	tenantID,
	subdomain,
	producerApplicationID,
	producerEnvironment,
	producerMicroserviceID,
	publicStreamID,
	partitionID,
	producerTenantID string) error {
	return nil
}

func AddConsumerTenantWithCustomDatabaseNameAndCopiedCustomSSL(
	applicationID,
	microserviceID,
	environment,
	tenantID,
	subdomain,
	producerApplicationID,
	producerEnvironment,
	producerMicroserviceID,
	publicStreamID,
	partitionID,
	producerTenantID,
	databaseName,
	sourceSSL,
	targetSSL string) error {

	return nil
}

func updateIngress() {

}

func updateTenantsJSON() {

}

func updateResources() {

}
