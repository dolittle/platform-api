package storage_test

import (
	"encoding/json"

	"github.com/dolittle/platform-api/pkg/platform/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Customer Tenants", func() {
	It("GetCustomerTenantsByEnvironment", func() {
		item := []byte(`{
			"id": "6eb64bea-769f-2f47-8fd6-42acb369b5cb",
			"name": "Run1",
			"tenantId": "b305719a-3a9c-4473-aaaa-5c60dbcb2049",
			"tenantName": "Dry1",
			"environments": [
			 {
			  "name": "Dev",
			  "customerTenants": [
			   {
				"alias": "",
				"environment": "Dev",
				"customerTenantId": "445f8ea8-1a6f-40d7-b2fc-796dba92dc44",
				"hosts": [
				 {
				  "host": "cool-swirles.dolittle.cloud",
				  "secretName": "cool-swirles-certificate"
				 }
				],
				"microservicesRel": [
				 {
				  "microserviceId": "3e239ed1-5cda-4fc8-9005-4d521b48c226",
				  "hash": "3e239ed_445f8ea"
				 }
				]
			   }
			  ],
			  "welcomeMicroserviceID": "3e239ed1-5cda-4fc8-9005-4d521b48c226"
			 }
			],
			"status": {
			 "status": "finished:success",
			 "startedAt": "2022-02-10T11:08:32Z",
			 "finishedAt": "2022-02-10T11:10:15Z"
			}
		   }`)
		var application storage.JSONApplication
		json.Unmarshal(item, &application)

		customerTenants := storage.GetCustomerTenantsByEnvironment(application, "Dev")
		Expect(len(customerTenants)).To(Equal(1))
		Expect(customerTenants[0].CustomerTenantID).To(Equal("445f8ea8-1a6f-40d7-b2fc-796dba92dc44"))
		Expect(len(customerTenants[0].Hosts)).To(Equal(1))
		Expect(customerTenants[0].Hosts[0].Host).To(Equal("cool-swirles.dolittle.cloud"))
	})

	It("GetCustomerTenants", func() {
		item := []byte(`{
			"id": "6eb64bea-769f-2f47-8fd6-42acb369b5cb",
			"name": "Run1",
			"tenantId": "b305719a-3a9c-4473-aaaa-5c60dbcb2049",
			"tenantName": "Dry1",
			"environments": [
			 {
			  "name": "Dev",
			  "customerTenants": [
			   {
				"alias": "",
				"environment": "Dev",
				"customerTenantId": "445f8ea8-1a6f-40d7-b2fc-796dba92dc44",
				"hosts": [
				 {
				  "host": "cool-swirles.dolittle.cloud",
				  "secretName": "cool-swirles-certificate"
				 }
				],
				"microservicesRel": [
				 {
				  "microserviceId": "3e239ed1-5cda-4fc8-9005-4d521b48c226",
				  "hash": "3e239ed_445f8ea"
				 }
				]
			   }
			  ],
			  "welcomeMicroserviceID": "3e239ed1-5cda-4fc8-9005-4d521b48c226"
			 }
			],
			"status": {
			 "status": "finished:success",
			 "startedAt": "2022-02-10T11:08:32Z",
			 "finishedAt": "2022-02-10T11:10:15Z"
			}
		   }`)
		var application storage.JSONApplication
		json.Unmarshal(item, &application)

		customerTenants := storage.GetCustomerTenants(application)
		Expect(len(customerTenants)).To(Equal(1))
	})

	It("GetCustomerTenants", func() {
		item := []byte(`{
			"id": "6eb64bea-769f-2f47-8fd6-42acb369b5cb",
			"name": "Run1",
			"tenantId": "b305719a-3a9c-4473-aaaa-5c60dbcb2049",
			"tenantName": "Dry1",
			"environments": [
			 {
			  "name": "Dev",
			  "customerTenants": [
			   {
				"alias": "",
				"environment": "Dev",
				"customerTenantId": "445f8ea8-1a6f-40d7-b2fc-796dba92dc44",
				"hosts": [
				 {
				  "host": "cool-swirles.dolittle.cloud",
				  "secretName": "cool-swirles-certificate"
				 }
				],
				"microservicesRel": [
				 {
				  "microserviceId": "3e239ed1-5cda-4fc8-9005-4d521b48c226",
				  "hash": "3e239ed_445f8ea"
				 }
				]
			   }
			  ],
			  "welcomeMicroserviceID": "3e239ed1-5cda-4fc8-9005-4d521b48c226"
			 }
			],
			"status": {
			 "status": "finished:success",
			 "startedAt": "2022-02-10T11:08:32Z",
			 "finishedAt": "2022-02-10T11:10:15Z"
			}
		   }`)
		var application storage.JSONApplication
		json.Unmarshal(item, &application)

		customerTenants := storage.GetCustomerTenantsByEnvironment(application, "Food")
		Expect(len(customerTenants)).To(Equal(0))
	})
})
