package k8s_test

import (
	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/microservice/simple/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/welcome"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	networkingv1 "k8s.io/api/networking/v1"
)

var _ = Describe("Test resource", func() {

	It("Confirm when HeadPort is set in extra the port propagates thru deployment and service", func() {
		customerID := "fake-customer-123"
		customerName := "fake-customer-name"
		applicationID := "fake-application-123"
		applicationName := "fake-application-name"

		customerInfo := dolittleK8s.Tenant{
			Name: customerName,
			ID:   customerID,
		}

		applicationInfo := dolittleK8s.Application{
			Name: applicationName,
			ID:   applicationID,
		}

		ms := platform.HttpInputSimpleInfo{
			MicroserviceBase: platform.MicroserviceBase{
				Dolittle: platform.HttpInputDolittle{
					ApplicationID:  applicationInfo.ID,
					CustomerID:     customerInfo.ID,
					MicroserviceID: "",
				},
				Name: "Welcome",
				Kind: platform.MicroserviceKindSimple,
			},
			Extra: platform.HttpInputSimpleExtra{
				Headimage:    welcome.Image,
				Runtimeimage: "none",
				Ingress: platform.HttpInputSimpleIngress{
					Path:     "/fake",
					Pathtype: string(networkingv1.PathTypePrefix),
				},
			},
		}

		customerTenants := make([]platform.CustomerTenantInfo, 0)

		tests := []struct {
			headPort int32
		}{
			{
				headPort: 80,
			},
			{
				headPort: 1234,
			},
		}

		for _, test := range tests {
			ms.Extra.HeadPort = test.headPort
			resources := k8s.NewResources(true, "test", customerInfo, applicationInfo, customerTenants, ms)

			Expect(resources.Service.Spec.Ports[0].Name).To(Equal("http"), "If this changes, the ingress might be broken")
			Expect(resources.Service.Spec.Ports[0].Port).To(Equal(test.headPort))
			Expect(resources.Service.Spec.Ports[0].TargetPort.IntVal).To(Equal(test.headPort))

			Expect(resources.Deployment.Spec.Template.Spec.Containers[0].Ports[0].Name).To(Equal("http"), "If this changes, the ingress might be broken")
			Expect(resources.Deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).To(Equal(test.headPort))
		}

	})

})
