package k8s_test

import (
	"fmt"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/microservice/simple/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resources", func() {

	var (
		isProduction     bool
		applicationID    string
		customerID       string
		microserviceID   string
		customerTenantID string
		environment      string
		namespace        string
		tenant           dolittleK8s.Tenant
		application      dolittleK8s.Application
		customerTenants  []platform.CustomerTenantInfo
		input            platform.HttpInputSimpleInfo
	)

	BeforeEach(func() {
		isProduction = false
		applicationID = "249f2c1b-fb46-49ac-956c-c71678f6eb92"
		customerID = "e95331e6-13b8-4d42-98ac-bca6bce501d3"
		microserviceID = "e243e12b-f5ec-41b0-94ad-1b2e67331446"
		customerTenantID = "db90aa1d-57fc-4d6b-8578-07c9ad9d7301"
		environment = "Test"

		namespace = fmt.Sprintf("application-%s", applicationID)
		tenant = dolittleK8s.Tenant{
			Name: "Test-Name",
			ID:   customerID,
		}
		application = dolittleK8s.Application{
			Name: "Test-Application",
			ID:   applicationID,
		}
		customerTenants = []platform.CustomerTenantInfo{
			{
				Alias:            "Test-Alias",
				Environment:      environment,
				CustomerTenantID: customerTenantID,
				Hosts: []platform.CustomerTenantHost{
					{
						Host:       "test-host",
						SecretName: "test-secret",
					},
				},
				MicroservicesRel: []platform.CustomerTenantMicroserviceRel{
					{
						MicroserviceID: microserviceID,
						Hash:           fmt.Sprintf("%s_%s", microserviceID[0:7], customerTenantID),
					},
				},
			},
		}

		input = platform.HttpInputSimpleInfo{
			MicroserviceBase: platform.MicroserviceBase{
				Dolittle: platform.HttpInputDolittle{
					ApplicationID:  applicationID,
					CustomerID:     customerID,
					MicroserviceID: microserviceID,
				},
				Name:        "Test-Microservice",
				Kind:        platform.MicroserviceKindSimple,
				Environment: environment,
			},
			Extra: platform.HttpInputSimpleExtra{
				Headimage:    "test-image",
				Runtimeimage: "dolittle/runtime:7.7.1",
				Ingress: platform.HttpInputSimpleIngress{
					Path:     "/",
					Pathtype: "Prefix",
				},
				Ispublic: false,
			},
		}
	})

	When("creating private microservice resources", func() {
		BeforeEach(func() {
			input.Extra.Ispublic = false
		})

		It("should have the resources set", func() {
			resources := k8s.NewResources(isProduction, namespace, tenant, application, customerTenants, input)
			Expect(resources.Service).ToNot(BeNil())
			Expect(resources.Deployment).ToNot(BeNil())
			Expect(resources.DolittleConfig).ToNot(BeNil())
			Expect(resources.ConfigFiles).ToNot(BeNil())
			Expect(resources.ConfigEnvironmentVariables).ToNot(BeNil())
			Expect(resources.SecretEnvironmentVariables).ToNot(BeNil())
			Expect(resources.RbacPolicyRules).ToNot(BeNil())
		})
	})

	When("creating public microservice resources", func() {
		BeforeEach(func() {
			input.Extra.Ispublic = true
		})

		It("should have the resources set", func() {
			resources := k8s.NewPublicResources(isProduction, namespace, tenant, application, customerTenants, input)
			Expect(resources.Service).ToNot(BeNil())
			Expect(resources.Deployment).ToNot(BeNil())
			Expect(resources.DolittleConfig).ToNot(BeNil())
			Expect(resources.ConfigFiles).ToNot(BeNil())
			Expect(resources.ConfigEnvironmentVariables).ToNot(BeNil())
			Expect(resources.SecretEnvironmentVariables).ToNot(BeNil())
			Expect(resources.RbacPolicyRules).ToNot(BeNil())
		})

		It("should have an ingress and network policy set", func() {
			resources := k8s.NewPublicResources(isProduction, namespace, tenant, application, customerTenants, input)
			Expect(resources.NetworkPolicy).ToNot(BeNil())
			Expect(resources.Ingresses).ToNot(BeNil())
		})
	})
})
