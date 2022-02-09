package k8s_test

import (
	"github.com/dolittle/platform-api/pkg/platform/microservice/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Microservice Resources", func() {
	It("NewMicroservicePolicyRules", func() {

		name := "Order"
		environment := "Dev"
		resource := k8s.NewMicroservicePolicyRules(name, environment)

		Expect(resource[0].Verbs).To(Equal([]string{
			"get",
			"patch",
		}), "Confirm verbs")

		Expect(resource[0].ResourceNames).To(Equal([]string{
			"dev-order-env-variables",
			"dev-order-config-files",
		}), "Confirm configmaps")

		Expect(resource[1].ResourceNames).To(Equal([]string{
			"dev-order-secret-env-variables",
		}), "Confirm secrets")
	})
})
