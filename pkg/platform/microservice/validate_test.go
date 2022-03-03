package microservice_test

import (
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/microservice"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Validate microservice", func() {

	When("Check microservice ingress path", func() {
		It("Ingress path not found", func() {
			ingresses := []platform.Ingress{}
			found := microservice.CheckIfIngressPathInUseInEnvironment(ingresses, "Dev", "/")
			Expect(found).To(BeFalse())
		})

		It("Ingress path not found in Dev", func() {
			ingresses := []platform.Ingress{
				{
					Host:        "test",
					Environment: "Prod",
					Path:        "/",
				},
			}
			found := microservice.CheckIfIngressPathInUseInEnvironment(ingresses, "Dev", "/")
			Expect(found).To(BeFalse())
		})

		It("Ingress path found in Dev", func() {
			ingresses := []platform.Ingress{
				{
					Host:        "test",
					Environment: "Dev",
					Path:        "/",
				},
			}
			found := microservice.CheckIfIngressPathInUseInEnvironment(ingresses, "Dev", "/")
			Expect(found).To(BeTrue())
		})
	})

})
