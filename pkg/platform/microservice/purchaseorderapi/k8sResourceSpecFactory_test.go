package purchaseorderapi_test

import (
	"fmt"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	. "github.com/dolittle-entropy/platform-api/pkg/platform/microservice/purchaseorderapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

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
			rawDataLogName  string
			tenant          platform.TenantId
		)

		BeforeEach(func() {
			headImage = "some-head-image"
			runtimeImage = "some-runtime-image"
			k8sMicroservice = k8s.Microservice{
				Environment: "Prod",
				Name:        "ramanujan",
				Application: k8s.Application{
					Name: "einstein",
					ID:   "12345-789",
				},
			}
			tenant = "fd93dfc9-8c44-4db7-844f-c0fde955792a"
			rawDataLogName = "raw-data-log-123"
			result = factory.CreateAll(headImage, runtimeImage, k8sMicroservice, tenant, platform.HttpInputPurchaseOrderExtra{
				RawDataLogName: rawDataLogName,
			})
		})
		It("should set the correct head image", func() {
			Expect(getContainerInDeployment(result.Deployment, "head").Image).To(Equal(headImage))
		})
		It("should set the correct runtime image", func() {
			Expect(getContainerInDeployment(result.Deployment, "runtime").Image).To(Equal(runtimeImage))
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
			Expect(result.ConfigEnvVariables.Data["TENANT"]).To(BeEquivalentTo(tenant))
		})
		It("should set SERVER_PORT to '8080'", func() {
			Expect(result.ConfigEnvVariables.Data["SERVER_PORT"]).To(Equal("8080"))
		})
		It("should set the correct NATS_CLUSTER_URL", func() {
			Expect(result.ConfigEnvVariables.Data["NATS_CLUSTER_URL"]).To(Equal(fmt.Sprintf("prod-%s-nats.application-12345-789.svc.cluster.local:4222", rawDataLogName)))
		})
		It("should set NATS_START_FROM_BEGINNING to 'false'", func() {
			Expect(result.ConfigEnvVariables.Data["NATS_START_FROM_BEGINNING"]).To(Equal("false"))
		})
		It("should set LOG_OUTPUT_FORMAT to 'json'", func() {
			Expect(result.ConfigEnvVariables.Data["LOG_OUTPUT_FORMAT"]).To(Equal("json"))
		})
	})
})

func getContainerInDeployment(deployment *appsV1.Deployment, containerName string) v1.Container {
	containers := deployment.Spec.Template.Spec.Containers
	var headContainer v1.Container
	for i := range containers {
		if containers[i].Name == containerName {
			headContainer = containers[i]
		}
	}
	return headContainer
}
