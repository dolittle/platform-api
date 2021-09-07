package purchaseorderapi_test

import (
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice"
	. "github.com/dolittle-entropy/platform-api/pkg/platform/microservice/purchaseorderapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
)

var _ = Describe("For k8sResourceSpecFactory", func() {
	var (
		factory K8sResourceSpecFactory
	)

	BeforeEach(func() {
		factory = NewK8sResourceSpecFactory()
	})

	Describe("when creating all resources", func() {
		var (
			headImage       string
			runtimeImage    string
			k8sMicroservice k8s.Microservice
			result          K8sResources
		)

		BeforeEach(func() {
			k8sMicroservice = k8s.Microservice{
				Environment: "Prod",
				Name:        "ramanujan",
				Application: k8s.Application{
					Name: "einstein",
					ID:   "12345-789",
				},
			}
			result = factory.CreateAll(headImage, runtimeImage, k8sMicroservice)
		})

		It("should set LOG_LEVEL to 'debug'", func() {
			Expect(result.ConfigEnvVariables.Data["LOG_LEVEL"]).To(Equal("debug"))
		})
		It("should set the correct DATABASE_READMODELS_URL", func() {
			Expect(result.ConfigEnvVariables.Data["DATABASE_READMODELS_URL"]).To(Equal("mongodb://prod-mongo.application-12345-789.svc.cluster.local:27017"))
		})
		It("should set the correct DATABASE_READMODELS_NAME", func() {
			Expect(result.ConfigEnvVariables.Data["DATABASE_READMODELS_NAME"]).To(Equal("einstein_prod_ramanujan_readmodels"))
		})
		It("should set NODE_ENV to 'production'", func() {
			Expect(result.ConfigEnvVariables.Data["NODE_ENV"]).To(Equal("production"))
		})
		It("should set TENANT to the todo-customer-tenant-id", func() {
			Expect(result.ConfigEnvVariables.Data["TENANT"]).To(Equal(microservice.TodoCustomersTenantID))
		})
		It("should set SERVER_PORT to '8080'", func() {
			Expect(result.ConfigEnvVariables.Data["SERVER_PORT"]).To(Equal("8080"))
		})
		It("should set the correct NATS_CLUSTER_URL", func() {
			Expect(result.ConfigEnvVariables.Data["NATS_CLUSTER_URL"]).To(Equal("prod-rawdatalogv1-nats.application-12345-789.svc.cluster.local:4222"))
		})
		It("should set NATS_START_FROM_BEGINNING to 'false'", func() {
			Expect(result.ConfigEnvVariables.Data["NATS_START_FROM_BEGINNING"]).To(Equal("false"))
		})
		It("should set LOG_OUTPUT_FORMAT to 'json'", func() {
			Expect(result.ConfigEnvVariables.Data["LOG_OUTPUT_FORMAT"]).To(Equal("json"))
		})
	})
})
