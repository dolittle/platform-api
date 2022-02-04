package storage_test

import (
	"encoding/json"

	"github.com/dolittle/platform-api/pkg/platform/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Customer Tenants", func() {
	It("GetCustomerTenantsByEnvironment", func() {
		item := []byte(`{"id":"fake-123","name":"Snow1","tenantId":"7df10b41-5fbd-4a4e-a2cf-27349d1f63ba","tenantName":"Test2","environments":[{"name":"Dev","tenantId":"7df10b41-5fbd-4a4e-a2cf-27349d1f63ba","applicationId":"fake-123","tenants":[],"ingresses":[],"customerTenants":[{"alias":"","environment":"Dev","customerTenantId":"445f8ea8-1a6f-40d7-b2fc-796dba92dc44","indexId":0,"ingress":{"host":"focused-feistel.dolittle.cloud","domainPrefix":"focused-feistel","secretName":"focused-feistel-certificate"},"microservicesRel":[{"microserviceID":"7d5d2a74-a6a1-4d52-8998-3970e445ee78","hash":"7d5d2a7_445f8ea"}],"runtime":{"databasePrefix":"7d5d2a7_445f8ea"}}],"customerTenantLastIndexID":0,"welcomeMicroserviceID":"7d5d2a74-a6a1-4d52-8998-3970e445ee78"},{"name":"Test","tenantId":"7df10b41-5fbd-4a4e-a2cf-27349d1f63ba","applicationId":"fake-123","tenants":[],"ingresses":[],"customerTenants":[{"alias":"","environment":"Test","customerTenantId":"445f8ea8-1a6f-40d7-b2fc-796dba92dc44","indexId":0,"ingress":{"host":"goofy-gates.dolittle.cloud","domainPrefix":"goofy-gates","secretName":"goofy-gates-certificate"},"microservicesRel":[{"microserviceID":"82febcb7-ec65-4ea1-84d8-0ccaa1b5365c","hash":"82febcb_445f8ea"}],"runtime":{"databasePrefix":"82febcb_445f8ea"}}],"customerTenantLastIndexID":0,"welcomeMicroserviceID":"82febcb7-ec65-4ea1-84d8-0ccaa1b5365c"}]}`)
		var application storage.JSONApplication
		json.Unmarshal(item, &application)

		customerTenants := storage.GetCustomerTenantsByEnvironment(application, "Dev")
		//b, _ := json.MarshalIndent(customerTenants, "", "  ")
		//output := string(b)
		//fmt.Println(output)
		Expect(len(customerTenants)).To(Equal(1))
	})

	It("GetCustomerTenants", func() {
		item := []byte(`{"id":"fake-123","name":"Snow1","tenantId":"7df10b41-5fbd-4a4e-a2cf-27349d1f63ba","tenantName":"Test2","environments":[{"name":"Dev","tenantId":"7df10b41-5fbd-4a4e-a2cf-27349d1f63ba","applicationId":"fake-123","tenants":[],"ingresses":[],"customerTenants":[{"alias":"","environment":"Dev","customerTenantId":"445f8ea8-1a6f-40d7-b2fc-796dba92dc44","indexId":0,"ingress":{"host":"focused-feistel.dolittle.cloud","domainPrefix":"focused-feistel","secretName":"focused-feistel-certificate"},"microservicesRel":[{"microserviceID":"7d5d2a74-a6a1-4d52-8998-3970e445ee78","hash":"7d5d2a7_445f8ea"}],"runtime":{"databasePrefix":"7d5d2a7_445f8ea"}}],"customerTenantLastIndexID":0,"welcomeMicroserviceID":"7d5d2a74-a6a1-4d52-8998-3970e445ee78"},{"name":"Test","tenantId":"7df10b41-5fbd-4a4e-a2cf-27349d1f63ba","applicationId":"fake-123","tenants":[],"ingresses":[],"customerTenants":[{"alias":"","environment":"Test","customerTenantId":"445f8ea8-1a6f-40d7-b2fc-796dba92dc44","indexId":0,"ingress":{"host":"goofy-gates.dolittle.cloud","domainPrefix":"goofy-gates","secretName":"goofy-gates-certificate"},"microservicesRel":[{"microserviceID":"82febcb7-ec65-4ea1-84d8-0ccaa1b5365c","hash":"82febcb_445f8ea"}],"runtime":{"databasePrefix":"82febcb_445f8ea"}}],"customerTenantLastIndexID":0,"welcomeMicroserviceID":"82febcb7-ec65-4ea1-84d8-0ccaa1b5365c"}]}`)
		var application storage.JSONApplication
		json.Unmarshal(item, &application)

		customerTenants := storage.GetCustomerTenants(application)

		//b, _ := json.MarshalIndent(customerTenants, "", "  ")
		//output := string(b)
		//fmt.Println(output)
		Expect(len(customerTenants)).To(Equal(2))
	})

	It("GetCustomerTenants", func() {
		item := []byte(`{"id":"fake-123","name":"Snow1","tenantId":"7df10b41-5fbd-4a4e-a2cf-27349d1f63ba","tenantName":"Test2","environments":[{"name":"Dev","tenantId":"7df10b41-5fbd-4a4e-a2cf-27349d1f63ba","applicationId":"fake-123","tenants":[],"ingresses":[],"customerTenants":[{"alias":"","environment":"Dev","customerTenantId":"445f8ea8-1a6f-40d7-b2fc-796dba92dc44","indexId":0,"ingress":{"host":"focused-feistel.dolittle.cloud","domainPrefix":"focused-feistel","secretName":"focused-feistel-certificate"},"microservicesRel":[{"microserviceID":"7d5d2a74-a6a1-4d52-8998-3970e445ee78","hash":"7d5d2a7_445f8ea"}],"runtime":{"databasePrefix":"7d5d2a7_445f8ea"}}],"customerTenantLastIndexID":0,"welcomeMicroserviceID":"7d5d2a74-a6a1-4d52-8998-3970e445ee78"},{"name":"Test","tenantId":"7df10b41-5fbd-4a4e-a2cf-27349d1f63ba","applicationId":"fake-123","tenants":[],"ingresses":[],"customerTenants":[{"alias":"","environment":"Test","customerTenantId":"445f8ea8-1a6f-40d7-b2fc-796dba92dc44","indexId":0,"ingress":{"host":"goofy-gates.dolittle.cloud","domainPrefix":"goofy-gates","secretName":"goofy-gates-certificate"},"microservicesRel":[{"microserviceID":"82febcb7-ec65-4ea1-84d8-0ccaa1b5365c","hash":"82febcb_445f8ea"}],"runtime":{"databasePrefix":"82febcb_445f8ea"}}],"customerTenantLastIndexID":0,"welcomeMicroserviceID":"82febcb7-ec65-4ea1-84d8-0ccaa1b5365c"}]}`)
		var application storage.JSONApplication
		json.Unmarshal(item, &application)

		customerTenants := storage.GetCustomerTenantsByEnvironment(application, "Food")

		//b, _ := json.MarshalIndent(customerTenants, "", "  ")
		//output := string(b)
		//fmt.Println(output)
		Expect(len(customerTenants)).To(Equal(0))
	})
})
