package k8s_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	networkingv1 "k8s.io/api/networking/v1"

	. "github.com/dolittle/platform-api/pkg/dolittle/k8s"
)

var _ = Describe("Ingress", func() {
	Describe("when creating a Ingress", func() {

		var (
			resource     *networkingv1.Ingress
			microservice Microservice
			isProduction bool
		)
		BeforeEach(func() {
			isProduction = false
			microservice = Microservice{}

		})

		It("should create a deployment with the correct ApiVersion", func() {
			resource = NewMicroserviceIngressWithEmptyRules(isProduction, microservice)
			Expect(resource.APIVersion).To(Equal("networking.k8s.io/v1"))
		})

		It("Default cluster-issuer", func() {
			resource = NewMicroserviceIngressWithEmptyRules(isProduction, microservice)
			Expect(resource.Annotations["cert-manager.io/cluster-issuer"]).To(Equal("letsencrypt-staging"))
		})

		It("when platform environment is prod using letsencrypt production", func() {
			isProduction = true
			resource = NewMicroserviceIngressWithEmptyRules(isProduction, microservice)
			Expect(resource.Annotations["cert-manager.io/cluster-issuer"]).To(Equal("letsencrypt-production"))
		})

	})
})
