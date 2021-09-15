package purchaseorderapi_test

import (
	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	. "github.com/dolittle-entropy/platform-api/pkg/platform/microservice/purchaseorderapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("For repo", func() {
	var (
		repo                   Repo
		k8sResource            K8sResource
		k8sResourceSpecFactory K8sResourceSpecFactory
		k8sClient              kubernetes.Interface
	)

	BeforeEach(func() {
		k8sClient = fake.NewSimpleClientset()
		k8sResourceSpecFactory = NewK8sResourceSpecFactory()
		k8sResource = NewK8sResource(k8sClient, k8sResourceSpecFactory)
		repo = NewRepo(k8sResource, k8sResourceSpecFactory, k8sClient)
	})

	Describe("when checking if purchase order api exists", func() {
		var (
			namespace   string
			tenant      k8s.Tenant
			application k8s.Application
			input       platform.HttpInputPurchaseOrderInfo
		)

		BeforeEach(func() {
			namespace = "some-namespace"
			tenant = k8s.Tenant{Name: "tenant", ID: "2ff0f068-a943-42b1-b704-2b9dec2574ed"}
			application = k8s.Application{Name: "application", ID: "810bf759-f276-4add-b765-62cfa1769c50"}
			input = platform.HttpInputPurchaseOrderInfo{}
		})
		Describe("and it does exist", func() {
			existsResult, errResult := repo.Exists(namespace, tenant, application, input)
		})
	})
})
