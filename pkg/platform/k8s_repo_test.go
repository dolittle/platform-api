package platform_test

import (
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("k8s repo test", func() {
	When("Working with the Ingress", func() {
		FIt("Get the Customer TenantID from nginx configuration", func() {
			sample := `proxy_set_header Tenant-ID "61838650-f8b7-412f-8e46-dc6165fc3dc4";`
			r, _ := regexp.Compile(`proxy_set_header Tenant-ID "(\S+)"`)

			matches := r.FindStringSubmatch(sample)
			Expect(len(matches)).To(Equal(2))
			Expect(matches[1]).To(Equal("61838650-f8b7-412f-8e46-dc6165fc3dc4"))
		})
	})
})
