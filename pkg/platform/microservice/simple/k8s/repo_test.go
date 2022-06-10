package k8s_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/testing"

	pkgK8s "github.com/dolittle/platform-api/pkg/k8s"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/simple"
	"github.com/dolittle/platform-api/pkg/platform/microservice/simple/k8s"
)

var _ = Describe("Repo", func() {

	var (
		clientSet              *fake.Clientset
		config                 *rest.Config
		k8sDolittleRepo        platformK8s.K8sRepo
		k8sRepoV2              pkgK8s.Repo
		repo                   simple.Repo
		logger                 *logrus.Logger
		consumerMicroserviceID string
		producerMicroserviceID string
		// getConsumerMicroservice *appsv1.Deployment
		environment            string
		consumerApplicationID  string
		producerApplicationID  string
		producerNamespaceError error
		consumerCustomerID     string
		producerCustomerID     string
		consumerTenantID       string
		producerTenantID       string
		publicStream           string
		partition              string
		err                    error
	)

	BeforeEach(func() {
		clientSet = &fake.Clientset{}
		config = &rest.Config{}
		logger, _ = logrusTest.NewNullLogger()
		k8sDolittleRepo = platformK8s.NewK8sRepo(clientSet, config, logger)
		k8sRepoV2 = pkgK8s.NewRepo(clientSet, logger)
		repo = k8s.NewSimpleRepo(clientSet, k8sDolittleRepo, k8sRepoV2, false)

		consumerMicroserviceID = "9fda2a06-01ec-4a77-b589-dac206a6be7c"
		producerMicroserviceID = "adbb5d8c-ec55-42a0-acc7-13a6b14f3c73"

		consumerApplicationID = "9bbda058-c59b-4362-94ce-40687b678302"
		producerApplicationID = "9bbda058-c59b-4362-94ce-40687b678302"
		producerNamespaceError = nil

		consumerCustomerID = "de87265a-af31-4fd5-b64f-8d8679858473"
		producerCustomerID = "de87265a-af31-4fd5-b64f-8d8679858473"

		consumerTenantID = "eed821d3-be32-4a2f-9a83-8f4808866ddb"
		producerTenantID = "2086ebc8-9be1-4300-a9d0-4acc8bb80781"

		partition = "00000000-0000-0000-0000-000000000000"
		publicStream = "18340123-2b68-4667-9190-460f1f3d9408"
	})

	Describe("Adding an event horizon subscription", func() {

		JustBeforeEach(func() {
			clientSet.AddReactor("get", "namespaces", func(action testing.Action) (bool, runtime.Object, error) {
				getAction := action.(testing.GetAction)
				getNamespace := getAction.GetName()
				// if strings.HasSuffix(getNamespace, consumerApplicationID) {
				// 	namespace := &corev1.Namespace{
				// 		ObjectMeta: metav1.ObjectMeta{
				// 			Name:      fmt.Sprintf("application-%s", consumerApplicationID),
				// 			Namespace: fmt.Sprintf("application-%s", consumerApplicationID),
				// 			Annotations: map[string]string{
				// 				"dolittle.io/tenant-id": consumerCustomerID,
				// 			},
				// 		},
				// 	}
				// 	return true, namespace, nil
				// }
				if strings.HasSuffix(getNamespace, producerApplicationID) {
					namespace := &corev1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("application-%s", producerApplicationID),
							Namespace: fmt.Sprintf("application-%s", producerApplicationID),
							Annotations: map[string]string{
								"dolittle.io/tenant-id": producerCustomerID,
							},
						},
					}
					return true, namespace, producerNamespaceError
				}

				return true, nil, nil
			})

			// clientSet.AddReactor("list", "deployments", func(action testing.Action) (bool, runtime.Object, error) {
			// 	filters := action.(testing.ListAction).GetListRestrictions()

			// 	deploymentList := &appsv1.DeploymentList{
			// 		Items: []appsv1.Deployment{
			// 			{
			// 				ObjectMeta: metav1.ObjectMeta{
			// 					Name: "consumertest",
			// 					Labels: map[string]string{
			// 						"tenant":       "fake-tenant",
			// 						"application":  "fake-application",
			// 						"environment":  environment,
			// 						"microservice": "fake-microservice",
			// 					},
			// 					Annotations: map[string]string{
			// 						"dolittle.io/microservice-id": microserviceID,
			// 					},
			// 				},
			// 			},
			// 		},
			// 	}

			// 	return true, nil, nil
			// })
			err = repo.SubscribeToAnotherApplication(consumerCustomerID, consumerApplicationID, environment, consumerMicroserviceID, consumerTenantID, producerMicroserviceID, producerTenantID, publicStream, partition, producerApplicationID, environment)
		})

		When("the consumer and producer aren't owned by the same customer", func() {
			BeforeEach(func() {
				producerCustomerID = "im a different id than the consumer"
			})

			It("should fail", func() {
				Expect(err).ToNot(BeNil())
			})
		})

		When("the consumer and producer are in different applications", func() {
			BeforeEach(func() {
				// a different applicationID
				producerApplicationID = "587a9e21-9ab9-4955-812e-22c86bd52dcf"
			})

			It("should not fail", func() {
				Expect(err).To(BeNil())
			})

			It("should create a networkpolicy between the microservices if it doesn't exist", func() {

			})
		})
	})
})
