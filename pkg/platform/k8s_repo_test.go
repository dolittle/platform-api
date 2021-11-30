package platform_test

import (
	"github.com/dolittle/platform-api/pkg/platform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("k8s repo test", func() {
	When("Working with the Ingress", func() {
		When("Get the Customer TenantID from nginx configuration", func() {
			It("Not found", func() {
				sample := `nothing`
				expect := ""
				customerTenantID := platform.GetCustomerTenantIDFromNginxConfigurationSnippet(sample)
				Expect(customerTenantID).To(Equal(expect))

			})
			It("Found", func() {
				sample := `proxy_set_header Tenant-ID "61838650-f8b7-412f-8e46-dc6165fc3dc4";`
				expect := "61838650-f8b7-412f-8e46-dc6165fc3dc4"
				customerTenantID := platform.GetCustomerTenantIDFromNginxConfigurationSnippet(sample)
				Expect(customerTenantID).To(Equal(expect))
			})
		})
	})
})
