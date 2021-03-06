package purchaseorderapi_test

import (
	"github.com/dolittle/platform-api/pkg/platform"
	. "github.com/dolittle/platform-api/pkg/platform/microservice/purchaseorderapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	"github.com/dolittle/platform-api/pkg/dolittle/k8s"
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
			headImage        string
			runtimeImage     string
			k8sMicroservice  k8s.Microservice
			result           K8sResources
			rawDataLogName   string
			customerTenantID string
			customerTenants  []platform.CustomerTenantInfo
		)

		BeforeEach(func() {
			headImage = "some-head-image"
			runtimeImage = "some-runtime-image"
			k8sMicroservice = k8s.Microservice{
				ID:          "98e31294-98b8-4dba-88d8-916e3851f018",
				Environment: "Prod",
				Name:        "ramanujan",
				Application: k8s.Application{
					Name: "einstein",
					ID:   "12345-789",
				},
			}
			customerTenantID = "fd93dfc9-8c44-4db7-844f-c0fde955792a"
			rawDataLogName = "raw-data-log-123"

			customerTenants = []platform.CustomerTenantInfo{
				{
					CustomerTenantID: customerTenantID,
					Hosts: []platform.CustomerTenantHost{
						{
							Host:       "fake-prefix.fake-host",
							SecretName: "fake-prefix",
						},
					},
				},
			}

			result = factory.CreateAll(headImage, runtimeImage, k8sMicroservice, customerTenants, platform.HttpInputPurchaseOrderExtra{
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
			Expect(result.ConfigEnvVariables.Data["DATABASE_READMODELS_NAME"]).To(Equal("98e3129_fd93dfc_readmodels"))
		})
		It("should set NODE_ENV to 'production'", func() {
			Expect(result.ConfigEnvVariables.Data["NODE_ENV"]).To(Equal("production"))
		})

		It("should set TENANT to the todo-customer-tenant-id", func() {
			Expect(result.ConfigEnvVariables.Data["TENANT"]).To(BeEquivalentTo(customerTenants[0].CustomerTenantID))
		})
		It("should set SERVER_PORT to '8080'", func() {
			Expect(result.ConfigEnvVariables.Data["SERVER_PORT"]).To(Equal("80"))
		})
		It("should set the correct NATS_CLUSTER_URL", func() {
			Expect(result.ConfigEnvVariables.Data["NATS_CLUSTER_URL"]).To(Equal("prod-nats.application-12345-789.svc.cluster.local:4222"))
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
