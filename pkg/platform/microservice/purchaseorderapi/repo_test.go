package purchaseorderapi_test

import (
	"fmt"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	. "github.com/dolittle-entropy/platform-api/pkg/platform/microservice/purchaseorderapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var fakeClient *fake.Clientset = fake.NewSimpleClientset()
var _ = Describe("For repo", func() {
	var (
		repo                   Repo
		k8sResource            K8sResource
		k8sResourceSpecFactory K8sResourceSpecFactory
		k8sClient              kubernetes.Interface
	)
	k8sClient = fakeClient

	BeforeEach(func() {
		k8sResourceSpecFactory = NewK8sResourceSpecFactory()
		k8sResource = NewK8sResource(k8sClient, k8sResourceSpecFactory)
		repo = NewRepo(k8sResource, k8sResourceSpecFactory, k8sClient)
	})

	Describe("when checking if purchase order api exists", func() {
		var (
			namespace    string
			tenant       k8s.Tenant
			application  k8s.Application
			input        platform.HttpInputPurchaseOrderInfo
			existsResult bool
			errResult    error
		)

		BeforeEach(func() {
			input = basicMicroservice
			namespace = fmt.Sprintf("application-%s", input.Dolittle.ApplicationID)
			tenant = k8s.Tenant{Name: "tenant-name", ID: basicMicroservice.Dolittle.TenantID}
			application = k8s.Application{Name: "application-name", ID: basicMicroservice.Dolittle.ApplicationID}
		})
		Describe("and it does exist", func() {
			BeforeEach(func() {
				fakeClient = clientWithDeployment(input.Name, input, tenant, application)
				existsResult, errResult = repo.Exists(namespace, tenant, application, input)
			})
			It("should not fail", func() {
				Expect(errResult).To(BeNil())
			})
			It("should not exist", func() {
				Expect(existsResult).To(BeTrue())
			})
		})
	})
})

var basicMicroservice platform.HttpInputPurchaseOrderInfo = platform.HttpInputPurchaseOrderInfo{
	MicroserviceBase: platform.MicroserviceBase{
		Dolittle: platform.HttpInputDolittle{
			ApplicationID:  "c1e08289-be4b-4557-9457-5de90e0ea54a",
			TenantID:       "67dcf38f-16e4-4b57-bff5-707cff3233ec",
			MicroserviceID: "ef97a13b-2597-42a3-9fcb-161add2264c7",
		},
		Name:        "some-name",
		Kind:        platform.MicroserviceKindPurchaseOrderAPI,
		Environment: "some-env",
	},
	Extra: platform.HttpInputPurchaseOrderExtra{
		RawDataLogName: "raw-data-log-name",
		Headimage:      "head-image",
		Runtimeimage:   "runtime-image",
		Webhooks:       []platform.RawDataLogIngestorWebhookConfig{},
	},
}

func clientWithDeployment(microserviceName string, input platform.HttpInputPurchaseOrderInfo, tenant k8s.Tenant, application k8s.Application) *fake.Clientset {
	return fake.NewSimpleClientset(k8s.NewDeployment(k8s.Microservice{
		ID:          input.Dolittle.MicroserviceID,
		Name:        microserviceName,
		Tenant:      tenant,
		Application: application,
		Environment: input.Environment,
		ResourceID:  "d3c7524d-a51a-4bbd-9a5e-bdf3abbd143c",
		Kind:        platform.MicroserviceKindPurchaseOrderAPI,
	}, input.Extra.Headimage, input.Extra.Runtimeimage))
}
