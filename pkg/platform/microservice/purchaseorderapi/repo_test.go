package purchaseorderapi_test

import (
	"fmt"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	. "github.com/dolittle-entropy/platform-api/pkg/platform/microservice/purchaseorderapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("For repo", func() {
	var (
		repo Repo
	)

	Describe("when checking if purchase order api exists", func() {
		var (
			namespace      string
			tenant         k8s.Tenant
			application    k8s.Application
			input          platform.HttpInputPurchaseOrderInfo
			deploymentInfo fakeDeploymentInfo
			existsResult   bool
			errResult      error
		)

		BeforeEach(func() {
			input = microservice
			namespace = fmt.Sprintf("application-%s", input.Dolittle.ApplicationID)
			tenant = k8s.Tenant{Name: "tenant-name", ID: microservice.Dolittle.TenantID}
			application = k8s.Application{Name: "application-name", ID: microservice.Dolittle.ApplicationID}
			deploymentInfo = fakeDeploymentInfo{
				input:       input,
				tenant:      tenant,
				application: application,
			}
		})
		Describe("and there is only one deployment in the namespace", func() {
			Describe("and it does exist", func() {
				BeforeEach(func() {
					deploymentInfo.deployedMicroserviceName = input.Name
					repo = createRepoWithClient(fake.NewSimpleClientset(createDeployment(deploymentInfo)))
					existsResult, errResult = repo.Exists(namespace, tenant, application, input)
				})
				It("should not fail", func() {
					Expect(errResult).To(BeNil())
				})
				It("should exist", func() {
					Expect(existsResult).To(BeTrue())
				})
			})
			Describe("and it does not exist", func() {
				BeforeEach(func() {
					deploymentInfo.deployedMicroserviceName = "some-other-ms"
					repo = createRepoWithClient(fake.NewSimpleClientset(createDeployment(deploymentInfo)))
					existsResult, errResult = repo.Exists(namespace, tenant, application, input)
				})
				It("should not fail", func() {
					Expect(errResult).To(BeNil())
				})
				It("should not exist", func() {
					Expect(existsResult).To(BeFalse())
				})
			})
		})
		Describe("and there are multiple deployments in the namespace", func() {
			Describe("and it does exist", func() {
				var (
					firstDeployment  *v1.Deployment
					secondDeployment *v1.Deployment
				)
				BeforeEach(func() {
					deploymentInfo.deployedMicroserviceName = input.Name
					firstDeployment = createDeployment(deploymentInfo)
					deploymentInfo.deployedMicroserviceName = "some-other-ms"
					secondDeployment = createDeployment(deploymentInfo)
					repo = createRepoWithClient(fake.NewSimpleClientset(firstDeployment, secondDeployment))
					existsResult, errResult = repo.Exists(namespace, tenant, application, input)
				})
				It("should not fail", func() {
					Expect(errResult).To(BeNil())
				})
				It("should exist", func() {
					Expect(existsResult).To(BeTrue())
				})
			})
			Describe("and it does not exist", func() {
				var (
					firstDeployment  *v1.Deployment
					secondDeployment *v1.Deployment
				)
				BeforeEach(func() {
					deploymentInfo.deployedMicroserviceName = "some-ms"
					firstDeployment = createDeployment(deploymentInfo)
					deploymentInfo.deployedMicroserviceName = "some-other-ms"
					secondDeployment = createDeployment(deploymentInfo)
					repo = createRepoWithClient(fake.NewSimpleClientset(firstDeployment, secondDeployment))
					existsResult, errResult = repo.Exists(namespace, tenant, application, input)
				})
				It("should not fail", func() {
					Expect(errResult).To(BeNil())
				})
				It("should not exist", func() {
					Expect(existsResult).To(BeFalse())
				})
			})
		})
	})
})

type fakeDeploymentInfo struct {
	deployedMicroserviceName string
	input                    platform.HttpInputPurchaseOrderInfo
	tenant                   k8s.Tenant
	application              k8s.Application
}

var microservice platform.HttpInputPurchaseOrderInfo = platform.HttpInputPurchaseOrderInfo{
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

func createDeployment(info fakeDeploymentInfo) *v1.Deployment {
	return k8s.NewDeployment(k8s.Microservice{
		ID:          info.input.Dolittle.MicroserviceID,
		Name:        info.deployedMicroserviceName,
		Tenant:      info.tenant,
		Application: info.application,
		Environment: info.input.Environment,
		ResourceID:  "d3c7524d-a51a-4bbd-9a5e-bdf3abbd143c",
		Kind:        platform.MicroserviceKindPurchaseOrderAPI,
	}, info.input.Extra.Headimage, info.input.Extra.Runtimeimage)
}

func createRepoWithClient(client kubernetes.Interface) Repo {
	k8sResourceSpecFactory := NewK8sResourceSpecFactory()
	k8sResource := NewK8sResource(client, k8sResourceSpecFactory)
	return NewRepo(k8sResource, k8sResourceSpecFactory, client)
}
