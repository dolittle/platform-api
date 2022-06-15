package k8s_test

import (
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/testing"

	mockPkgK8s "github.com/dolittle/platform-api/mocks/pkg/k8s"
	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/simple"
	"github.com/dolittle/platform-api/pkg/platform/microservice/simple/k8s"
)

var _ = Describe("Repo", func() {

	var (
		clientSet                *fake.Clientset
		config                   *rest.Config
		k8sDolittleRepo          platformK8s.K8sRepo
		mockK8sRepoV2            *mockPkgK8s.Repo
		repo                     simple.Repo
		logger                   *logrus.Logger
		consumerMicroserviceID   string
		producerMicroserviceID   string
		consumerEnvironment      string
		producerEnvironment      string
		consumerApplicationID    string
		producerApplicationID    string
		producerNamespaceError   error
		consumerCustomerID       string
		producerCustomerID       string
		consumerTenantID         string
		producerTenantID         string
		publicStream             string
		partition                string
		scope                    string
		err                      error
		consumerConfigMap        *corev1.ConfigMap
		producerConfigMap        *corev1.ConfigMap
		consumerDeployment       *appsv1.Deployment
		producerDeployment       *appsv1.Deployment
		consumerUpdatedConfigMap *corev1.ConfigMap
		updatedMicroservices     dolittleK8s.MicroserviceMicroservices
		updatedEventHorizons     dolittleK8s.MicroserviceEventHorizons
		producerUpdatedConfigMap *corev1.ConfigMap
		updatedConsents          dolittleK8s.MicroserviceEventHorizonConsents
		producerService          *corev1.Service
		consumerMicroservices    dolittleK8s.MicroserviceMicroservices
		consumerEventHorizons    dolittleK8s.MicroserviceEventHorizons
		producerNamespace        string
		consumerNamespace        string
		producerConsents         dolittleK8s.MicroserviceEventHorizonConsents
		networkPolicy            *networkingv1.NetworkPolicy
		consumerName             string
		producerName             string
		consumerLabels           map[string]string
		producerLabels           map[string]string
		hasNetworkPolicy         bool
	)

	BeforeEach(func() {
		clientSet = &fake.Clientset{}
		config = &rest.Config{}
		logger, _ = logrusTest.NewNullLogger()
		k8sDolittleRepo = platformK8s.NewK8sRepo(clientSet, config, logger)
		mockK8sRepoV2 = new(mockPkgK8s.Repo)
		repo = k8s.NewSimpleRepo(clientSet, k8sDolittleRepo, mockK8sRepoV2, false)

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
		scope = "264d1d94-eda1-46ee-ae3d-7057b798b7a1"

		consumerEnvironment = "test"
		producerEnvironment = "dev"
		consumerName = "consumer"
		producerName = "producer"
	})

	Describe("Adding an event horizon subscription", func() {

		JustBeforeEach(func() {

			producerNamespace = fmt.Sprintf("application-%s", producerApplicationID)
			consumerNamespace = fmt.Sprintf("application-%s", consumerApplicationID)

			consumerAnnotations := map[string]string{
				"dolittle.io/tenant-id":       consumerCustomerID,
				"dolittle.io/microservice-id": consumerMicroserviceID,
				"dolittle.io/application-id":  consumerApplicationID,
			}
			producerAnnotations := map[string]string{
				"dolittle.io/tenant-id":       producerCustomerID,
				"dolittle.io/microservice-id": producerMicroserviceID,
				"dolittle.io/application-id":  producerApplicationID,
			}

			consumerLabels = map[string]string{
				"environment":  consumerEnvironment,
				"application":  "ConsumerApplication",
				"tenant":       "HorizonCustomer",
				"microservice": "Consumer",
			}
			producerLabels = map[string]string{
				"environment":  producerEnvironment,
				"application":  "ConsumerApplication",
				"tenant":       "HorizonCustomer",
				"microservice": "Producer",
			}

			producerDeployment = &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:        fmt.Sprintf("%s-%s", producerEnvironment, producerName),
					Namespace:   producerNamespace,
					Annotations: producerAnnotations,
					Labels:      producerLabels,
				},
			}
			mockK8sRepoV2.On("GetDeployment", producerNamespace, producerEnvironment, producerMicroserviceID).
				Return(*producerDeployment, nil)

			consumerDeployment = &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:        fmt.Sprintf("%s-%s", consumerEnvironment, consumerName),
					Namespace:   consumerNamespace,
					Annotations: consumerAnnotations,
					Labels:      consumerLabels,
				},
			}
			mockK8sRepoV2.On("GetDeployment", consumerNamespace, consumerEnvironment, consumerMicroserviceID).
				Return(*consumerDeployment, nil)

			producerService = &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:        fmt.Sprintf("%s-%s", producerEnvironment, producerName),
					Namespace:   producerNamespace,
					Labels:      producerLabels,
					Annotations: producerAnnotations,
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
			b, _ = json.MarshalIndent(consumerEventHorizons, "", "  ")
			eventHorizonsJSON := string(b)

			consumerConfigMap = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:        fmt.Sprintf("%s-%s-dolittle", consumerEnvironment, consumerName),
					Namespace:   consumerNamespace,
					Annotations: consumerAnnotations,
					Labels:      consumerLabels,
				},
				Data: map[string]string{
					"microservices.json":  microservicesJSON,
					"event-horizons.json": eventHorizonsJSON,
				},
			}

			b, _ = json.MarshalIndent(producerConsents, "", "  ")
			consentsJSON := string(b)

			producerConfigMap = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:        fmt.Sprintf("%s-%s-dolittle", producerEnvironment, producerName),
					Namespace:   producerNamespace,
					Annotations: producerAnnotations,
					Labels:      producerLabels,
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
							Name:      producerNamespace,
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
					json.Unmarshal([]byte(consumerUpdatedConfigMap.Data["event-horizons.json"]), &updatedEventHorizons)
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

			clientSet.AddReactor("list", "networkpolicies", func(action testing.Action) (bool, runtime.Object, error) {
				listAction := action.(testing.ListActionImpl)
				networkPolicyList := networkingv1.NetworkPolicyList{}

				if listAction.Namespace == producerNamespace {
					// control setting up the networkpolicy
					if hasNetworkPolicy {

						networkPolicyList.Items = append(networkPolicyList.Items, networkingv1.NetworkPolicy{
							ObjectMeta: metav1.ObjectMeta{
								Name:        fmt.Sprintf("%s-%s-event-horizons", producerEnvironment, producerName),
								Namespace:   producerNamespace,
								Annotations: producerAnnotations,
								Labels:      producerLabels,
							},
							Spec: networkingv1.NetworkPolicySpec{
								PodSelector: metav1.LabelSelector{
									MatchLabels: producerLabels,
								},
								PolicyTypes: []networkingv1.PolicyType{"Ingress"},
								Ingress: []networkingv1.NetworkPolicyIngressRule{
									{
										From: []networkingv1.NetworkPolicyPeer{
											{
												NamespaceSelector: &metav1.LabelSelector{
													MatchLabels: map[string]string{
														"tenant":      consumerLabels["tenant"],
														"application": consumerLabels["appliation"],
													},
												},
												PodSelector: &metav1.LabelSelector{
													MatchLabels: map[string]string{
														"environment":  consumerLabels["environment"],
														"microservice": consumerLabels["microservice"],
													},
												},
											},
										},
									},
								},
							},
						})
					}

				}
				return true, &networkPolicyList, nil
			})

			clientSet.AddReactor("create", "networkpolicies", func(action testing.Action) (bool, runtime.Object, error) {
				createAction := action.(testing.CreateActionImpl)
				networkPolicy = createAction.GetObject().(*networkingv1.NetworkPolicy)
				return true, networkPolicy, nil
			})

			consumerDeployment = &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:        fmt.Sprintf("%s-%s", consumerEnvironment, consumerName),
					Namespace:   consumerNamespace,
					Annotations: consumerAnnotations,
					Labels:      consumerLabels,
				},
			}

			clientSet.AddReactor("list", "deployments", func(action testing.Action) (bool, runtime.Object, error) {
				listAction := action.(testing.ListActionImpl)
				deploymentList := appsv1.DeploymentList{}

				if listAction.Namespace == producerNamespace {
					deploymentList.Items = append(deploymentList.Items, *producerDeployment)
				}
				if listAction.Namespace == consumerNamespace {
					deploymentList.Items = append(deploymentList.Items, *consumerDeployment)
				}
				return true, &deploymentList, nil
			})

			err = repo.SubscribeToAnotherApplication(consumerCustomerID, consumerApplicationID, consumerEnvironment, consumerMicroserviceID, consumerTenantID, producerMicroserviceID, producerTenantID, publicStream, partition, producerApplicationID, producerEnvironment, scope)
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

			Context("and they are owned by the same customer", func() {
				Context("and the producer doesn't have a networkpolicy for the consumer", func() {
					BeforeEach(func() {
						hasNetworkPolicy = false
					})

					It("should create a networkpolicy between the microservices if it doesn't exist", func() {
						Expect(networkPolicy).ToNot(BeNil())
					})

					It("should create a networkpolicy with the correct podselector labels", func() {
						Expect(networkPolicy.Spec.PodSelector.MatchLabels).To(Equal(producerLabels))
					})

					It("should create a networkpolicy with the correct ingress from selector labels", func() {
						Expect(networkPolicy.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels["tenant"]).To(Equal(consumerLabels["tenant"]))
						Expect(networkPolicy.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels["application"]).To(Equal(consumerLabels["application"]))
						Expect(networkPolicy.Spec.Ingress[0].From[0].PodSelector.MatchLabels["environment"]).To(Equal(consumerLabels["environment"]))
						Expect(networkPolicy.Spec.Ingress[0].From[0].PodSelector.MatchLabels["microservice"]).To(Equal(consumerLabels["microservice"]))
					})
				})
				Context("and the producer has a networkpolicy for the consumer", func() {
					BeforeEach(func() {
						hasNetworkPolicy = true
					})

					It("should not fail", func() {
						Expect(err).To(BeNil())
					})

					It("should have gotten the producers microservice", func() {
						mockK8sRepoV2.AssertCalled(GinkgoT(), "GetDeployment", producerNamespace, producerEnvironment, producerMicroserviceID)
					})
					It("should have gotten the consumers microservice", func() {
						mockK8sRepoV2.AssertCalled(GinkgoT(), "GetDeployment", consumerNamespace, consumerEnvironment, consumerMicroserviceID)
					})

					It("should update the consumers microservices.json with the producers tenant", func() {
						Expect(updatedMicroservices[producerMicroserviceID]).ToNot(BeNil())
					})

					It("should update the consumers microservices.json with the producers full hostname and port", func() {
						hostname := fmt.Sprintf("%s-%s.svc.cluster.local", producerService.Name, producerNamespace)
						Expect(updatedMicroservices[producerMicroserviceID].Host).To(Equal(hostname))
						Expect(updatedMicroservices[producerMicroserviceID].Port).To(Equal(producerService.Spec.Ports[1].Port))
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

					It("Should update the consumers event-horizons.json", func() {
						Expect(updatedEventHorizons[consumerTenantID]).ToNot(BeEmpty())
						Expect(len(updatedEventHorizons[consumerTenantID])).To(Equal(1))
						Expect(updatedEventHorizons[consumerTenantID][0].Scope).To(Equal(scope))
						Expect(updatedEventHorizons[consumerTenantID][0].Microservice).To(Equal(producerMicroserviceID))
						Expect(updatedEventHorizons[consumerTenantID][0].Tenant).To(Equal(producerTenantID))
						Expect(updatedEventHorizons[consumerTenantID][0].Stream).To(Equal(publicStream))
						Expect(updatedEventHorizons[consumerTenantID][0].Partition).To(Equal(partition))
					})
				})
			})

			Context("and they are owned by different customers", func() {
				BeforeEach(func() {
					producerCustomerID = "im not the consumers customer"
				})

				It("should fail", func() {
					Expect(err).ToNot(BeNil())
				})
			})
		})
	})
})
