package simple

import (
	"github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
)

type Repo interface {
	Create(namespace string, tenant k8s.Tenant, application k8s.Application, customerTenants []platform.CustomerTenantInfo, input platform.HttpInputSimpleInfo) error
	Delete(applicationID, environment, microserviceID string) error
	Subscribe(customerID, applicationID, environment, microserviceID, tenantID, producerMicroserviceID, producerTenantID, publicStream, partition string) error
	SubscribeToAnotherApplication(customerID, applicationID, environment, microserviceID, tenantID, producerMicroserviceID, producerTenantID, publicStream, partition, producerApplicationID, producerEnvironment, scope string) error
}
