package k8s_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	. "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/testutils"
)

var _ = Describe("Configmaps", func() {
	Describe("when creating a NewMicroserviceConfigmap", func() {
		var (
			microservice     Microservice
			customerTenantID string
			customerTenants  []platform.CustomerTenantInfo
			resource         *corev1.ConfigMap
			microserviceID   string
		)

		BeforeEach(func() {
			customerTenantID = "fake-customer-tenant-id-123"
			microserviceID = "c974b165-38d7-4745-9c62-f78fa615682a"
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
			microservice = Microservice{
				ID:          microserviceID,
				Name:        "LeliaKim",
				Environment: "AndreJensen",
				Tenant: Tenant{
					ID:   "4acf6e5b-6fb2-4a29-8073-3a79707ab558",
					Name: "JeanetteJohnston",
				},
				Application: Application{
					ID:   "cc142a0d-deac-4974-ada9-de6e21337dca",
					Name: "AlejandroRiley",
				},
			}
			// TODO do we test for dolittle resources?
			resource = NewMicroserviceConfigmap(microservice, customerTenants)
		})

		It("should create a configmap with the correct ApiVersion", func() {
			Expect(resource.APIVersion).To(Equal("v1"))
		})

		It("should create a configmap with the correct Kind", func() {
			Expect(resource.Kind).To(Equal("ConfigMap"))
		})

		It("should create a configmap with the correct Namespace", func() {
			Expect(resource.Namespace).To(Equal(fmt.Sprintf("application-%s", microservice.Application.ID)))
		})

		It("should create a configmap with the correct tenant-id annotation", func() {
			Expect(resource.Annotations["dolittle.io/tenant-id"]).To(Equal(microservice.Tenant.ID))
		})
		It("should create a configmap with the correct application-id annotation", func() {
			Expect(resource.Annotations["dolittle.io/application-id"]).To(Equal(microservice.Application.ID))
		})
		It("should create a configmap with the correct microservice-id annotation", func() {
			Expect(resource.Annotations["dolittle.io/microservice-id"]).To(Equal(microservice.ID))
		})

		It("should create a configmap with the correct tenant label", func() {
			Expect(resource.Labels["tenant"]).To(Equal(microservice.Tenant.Name))
		})
		It("should create a configmap with the correct application label", func() {
			Expect(resource.Labels["application"]).To(Equal(microservice.Application.Name))
		})
		It("should create a configmap with the correct environment label", func() {
			Expect(resource.Labels["environment"]).To(Equal(microservice.Environment))
		})
		It("should create a configmap with the correct microservice label", func() {
			Expect(resource.Labels["microservice"]).To(Equal(microservice.Name))
		})

		It("should create a configmap with data attribute metrics.json", func() {
			want := `{
				"port": 9700
			  }`
			testutils.CheckJSONPrettyPrint(resource.Data["metrics.json"], want)
		})
	})

	Describe("when creating a NewConfigFilesConfigmap", func() {
		var (
			microservice Microservice
			resource     *corev1.ConfigMap
		)

		BeforeEach(func() {
			microservice = Microservice{
				ID:          "c974b165-38d7-4745-9c62-f78fa615682a",
				Name:        "LeliaKim",
				Environment: "AndreJensen",
				Tenant: Tenant{
					ID:   "4acf6e5b-6fb2-4a29-8073-3a79707ab558",
					Name: "JeanetteJohnston",
				},
				Application: Application{
					ID:   "cc142a0d-deac-4974-ada9-de6e21337dca",
					Name: "AlejandroRiley",
				},
			}

			resource = NewConfigFilesConfigmap(microservice)
		})

		It("should create a configmap with the correct ApiVersion", func() {
			Expect(resource.APIVersion).To(Equal("v1"))
		})

		It("should create a configmap with the correct Kind", func() {
			Expect(resource.Kind).To(Equal("ConfigMap"))
		})

		It("should create a configmap with the correct Namespace", func() {
			Expect(resource.Namespace).To(Equal(fmt.Sprintf("application-%s", microservice.Application.ID)))
		})

		It("should create a configmap with the correct tenant-id annotation", func() {
			Expect(resource.Annotations["dolittle.io/tenant-id"]).To(Equal(microservice.Tenant.ID))
		})
		It("should create a configmap with the correct application-id annotation", func() {
			Expect(resource.Annotations["dolittle.io/application-id"]).To(Equal(microservice.Application.ID))
		})
		It("should create a configmap with the correct microservice-id annotation", func() {
			Expect(resource.Annotations["dolittle.io/microservice-id"]).To(Equal(microservice.ID))
		})

		It("should create a configmap with the correct tenant label", func() {
			Expect(resource.Labels["tenant"]).To(Equal(microservice.Tenant.Name))
		})
		It("should create a configmap with the correct application label", func() {
			Expect(resource.Labels["application"]).To(Equal(microservice.Application.Name))
		})
		It("should create a configmap with the correct environment label", func() {
			Expect(resource.Labels["environment"]).To(Equal(microservice.Environment))
		})
		It("should create a configmap with the correct microservice label", func() {
			Expect(resource.Labels["microservice"]).To(Equal(microservice.Name))
		})
	})
})
