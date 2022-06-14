package k8s_test

import (
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/testing"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
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
		environment              string
		consumerApplicationID    string
		producerApplicationID    string
		producerNamespaceError   error
		consumerCustomerID       string
		producerCustomerID       string
		consumerTenantID         string
		producerTenantID         string
		publicStream             string
		partition                string
		err                      error
		consumerConfigMap        *corev1.ConfigMap
		producerConfigMap        *corev1.ConfigMap
		consumerUpdatedConfigMap *corev1.ConfigMap
		updatedMicroservices     dolittleK8s.MicroserviceMicroservices
		producerUpdatedConfigMap *corev1.ConfigMap
		updatedConsents          dolittleK8s.MicroserviceEventHorizonConsents
		producerService          *corev1.Service
		consumerMicroservices    dolittleK8s.MicroserviceMicroservices
		producerNamespace        string
		consumerNamespace        string
		producerConsents         dolittleK8s.MicroserviceEventHorizonConsents
	)

	BeforeEach(func() {
		clientSet = &fake.Clientset{}
		config = &rest.Config{}
		logger, _ = logrusTest.NewNullLogger()
		k8sDolittleRepo = platformK8s.NewK8sRepo(clientSet, config, logger)
		k8sRepoV2 = pkgK8s.NewRepo(clientSet, logger)
		repo = k8s.NewSimpleRepo(clientSet, k8sDolittleRepo, k8sRepoV2, false)
		environment = "test"

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
		updatedMicroservices = dolittleK8s.MicroserviceMicroservices{}
		updatedConsents = dolittleK8s.MicroserviceEventHorizonConsents{}
	})

	Describe("Adding an event horizon subscription", func() {

		JustBeforeEach(func() {

			producerNamespace = fmt.Sprintf("application-%s", producerApplicationID)
			consumerNamespace = fmt.Sprintf("application-%s", consumerApplicationID)
			producerService = &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-producer-service",
					Namespace: producerNamespace,
					Labels: map[string]string{
						"environment": "Test",
					},
					Annotations: map[string]string{
						"dolittle.io/tenant-id":       producerCustomerID,
						"dolittle.io/microservice-id": producerMicroserviceID,
						"dolittle.io/application-id":  producerApplicationID,
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromString("http"),
						},
						{
							Name:       "runtime",
							Port:       50052,
							TargetPort: intstr.FromString("runtime"),
						},
					},
				},
			}

			b, _ := json.MarshalIndent(consumerMicroservices, "", "  ")
			microservicesJSON := string(b)

			consumerConfigMap = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-consumer-dolittle",
					Namespace: consumerNamespace,
					Annotations: map[string]string{
						"dolittle.io/tenant-id":       consumerCustomerID,
						"dolittle.io/microservice-id": consumerMicroserviceID,
						"dolittle.io/application-id":  consumerApplicationID,
					},
					Labels: map[string]string{
						"environment": "Test",
					},
				},
				Data: map[string]string{
					"microservices.json": microservicesJSON,
				},
			}

			b, _ = json.MarshalIndent(producerConsents, "", "  ")
			consentsJSON := string(b)

			producerConfigMap = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-producer-dolittle",
					Namespace: producerNamespace,
					Annotations: map[string]string{
						"dolittle.io/tenant-id":       producerCustomerID,
						"dolittle.io/microservice-id": producerMicroserviceID,
						"dolittle.io/application-id":  producerApplicationID,
					},
					Labels: map[string]string{
						"environment": "Test",
					},
				},
				Data: map[string]string{
					"event-horizon-consents.json": consentsJSON,
				},
			}

			clientSet.AddReactor("get", "namespaces", func(action testing.Action) (bool, runtime.Object, error) {
				getAction := action.(testing.GetAction)
				getNamespace := getAction.GetName()
				if strings.HasSuffix(getNamespace, producerApplicationID) {
					namespace := &corev1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name:      fmt.Sprintf("application-%s", producerApplicationID),
							Namespace: producerNamespace,
							Annotations: map[string]string{
								"dolittle.io/tenant-id": producerCustomerID,
							},
						},
					}
					return true, namespace, producerNamespaceError
				}

				return true, nil, nil
			})

			clientSet.AddReactor("list", "configmaps", func(action testing.Action) (bool, runtime.Object, error) {
				listAction := action.(testing.ListActionImpl)

				configMapList := corev1.ConfigMapList{}
				if listAction.Namespace == producerNamespace {
					configMapList.Items = append(configMapList.Items, *producerConfigMap)
				}
				if listAction.Namespace == consumerNamespace {
					configMapList.Items = append(configMapList.Items, *consumerConfigMap)
				}
				return true, &configMapList, nil
			})

			clientSet.AddReactor("update", "configmaps", func(action testing.Action) (bool, runtime.Object, error) {
				updateAction := action.(testing.UpdateAction)
				configmap := updateAction.GetObject().(*corev1.ConfigMap)
				if configmap.ObjectMeta.Annotations["dolittle.io/microservice-id"] == consumerMicroserviceID {
					consumerUpdatedConfigMap = updateAction.GetObject().(*corev1.ConfigMap)
					json.Unmarshal([]byte(consumerUpdatedConfigMap.Data["microservices.json"]), &updatedMicroservices)
					fmt.Println(updatedMicroservices)
					return true, consumerUpdatedConfigMap, nil
				}
				if configmap.ObjectMeta.Annotations["dolittle.io/microservice-id"] == producerMicroserviceID {
					producerUpdatedConfigMap = updateAction.GetObject().(*corev1.ConfigMap)
					json.Unmarshal([]byte(producerUpdatedConfigMap.Data["event-horizon-consents.json"]), &updatedConsents)
					return true, producerUpdatedConfigMap, nil
				}
				return true, nil, nil
			})

			clientSet.AddReactor("list", "services", func(action testing.Action) (bool, runtime.Object, error) {
				listAction := action.(testing.ListActionImpl)
				producerNamespace := fmt.Sprintf("application-%s", producerApplicationID)

				serviceList := corev1.ServiceList{}
				if listAction.Namespace == producerNamespace {
					serviceList = corev1.ServiceList{
						Items: []corev1.Service{
							*producerService,
						},
					}
				}
				return true, &serviceList, nil
			})

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

			It("should update the consumers microservices.json with the producers tenant", func() {
				Expect(updatedMicroservices[producerMicroserviceID]).ToNot(BeNil())
			})

			It("should update the consumers microservices.json with the producers full hostname and port", func() {
				hostname := fmt.Sprintf("%s-%s.svc.cluster.local", producerService.Name, producerNamespace)
				Expect(updatedMicroservices[producerMicroserviceID].Host).To(Equal(hostname))
				Expect(updatedMicroservices[producerMicroserviceID].Port).To(Equal(producerService.Spec.Ports[1].Port))
			})

			XIt("should create a networkpolicy between the microservices if it doesn't exist", func() {

			})

			It("should update the producers event-horizon-consents.json", func() {
				Expect(updatedConsents[producerTenantID]).ToNot(BeEmpty())
				Expect(len(updatedConsents[producerTenantID])).To(Equal(1))
				Expect(updatedConsents[producerTenantID][0].Microservice).To(Equal(consumerMicroserviceID))
				Expect(updatedConsents[producerTenantID][0].Tenant).To(Equal(consumerTenantID))
				Expect(updatedConsents[producerTenantID][0].Stream).To(Equal(publicStream))
				Expect(updatedConsents[producerTenantID][0].Partition).To(Equal(partition))
				Expect(updatedConsents[producerTenantID][0].Consent).ToNot(BeNil())
			})
		})
	})
})
