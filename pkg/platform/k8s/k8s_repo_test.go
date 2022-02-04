package k8s_test

import (
	"errors"

	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
	appsv1 "k8s.io/api/apps/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/testing"
)

var _ = Describe("k8s repo test", func() {
	When("Working with the Ingress", func() {
		When("Get the Customer TenantID from nginx configuration", func() {
			It("Not found", func() {
				sample := `nothing`
				expect := ""
				customerTenantID := platformK8s.GetCustomerTenantIDFromNginxConfigurationSnippet(sample)
				Expect(customerTenantID).To(Equal(expect))

			})
			It("Found", func() {
				sample := `proxy_set_header Tenant-ID "61838650-f8b7-412f-8e46-dc6165fc3dc4";`
				expect := "61838650-f8b7-412f-8e46-dc6165fc3dc4"
				customerTenantID := platformK8s.GetCustomerTenantIDFromNginxConfigurationSnippet(sample)
				Expect(customerTenantID).To(Equal(expect))
			})
		})

		When("GetIngressHTTPIngressPath", func() {
			var (
				applicationID  string
				environment    string
				microserviceID string
				want           error
				clientSet      *fake.Clientset
				config         *rest.Config
				k8sRepo        platformK8s.K8sRepo
				logger         *logrus.Logger
			)

			BeforeEach(func() {
				applicationID = "fake-application-123"
				environment = "Dev"
				microserviceID = "fake-microservice-123"
				want = errors.New("fail")
				clientSet = &fake.Clientset{}
				config = &rest.Config{}
				logger, _ = logrusTest.NewNullLogger()
				k8sRepo = platformK8s.NewK8sRepo(clientSet, config, logger.WithField("context", "k8s-repo"))

			})

			It("Error getting list", func() {
				clientSet.AddReactor("list", "ingresses", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					data := &networkingv1.IngressList{}
					return true, data, want
				})

				_, err := k8sRepo.GetIngressHTTPIngressPath(applicationID, environment, microserviceID)
				Expect(err).To(Equal(want))
			})

			It("When List is empty", func() {
				clientSet.AddReactor("list", "ingresses", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					filters := action.(testing.ListActionImpl).ListRestrictions
					Expect(filters.Labels.Matches(labels.Set{"environment": environment})).To(BeTrue())

					data := &networkingv1.IngressList{}
					return true, data, nil
				})

				items, err := k8sRepo.GetIngressHTTPIngressPath(applicationID, environment, microserviceID)
				Expect(err).To(BeNil())
				Expect(len(items)).To(Equal(0))
			})

			It("When List has no ingresses linked to this microservice", func() {
				clientSet.AddReactor("list", "ingresses", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					filters := action.(testing.ListActionImpl).ListRestrictions
					Expect(filters.Labels.Matches(labels.Set{"environment": environment})).To(BeTrue())
					className := "nginx"

					data := &networkingv1.IngressList{
						Items: []networkingv1.Ingress{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name: "ignore-1",
									Labels: map[string]string{
										"environment": environment,
									},
								},
								Spec: networkingv1.IngressSpec{
									IngressClassName: &className,
									Rules:            []networkingv1.IngressRule{},
								},
							},
						},
					}
					return true, data, nil
				})

				items, err := k8sRepo.GetIngressHTTPIngressPath(applicationID, environment, microserviceID)
				Expect(err).To(BeNil())
				Expect(len(items)).To(Equal(0))
			})

			When("Ingresses are found", func() {
				It("Only 1", func() {
					clientSet.AddReactor("list", "ingresses", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
						filters := action.(testing.ListActionImpl).ListRestrictions
						Expect(filters.Labels.Matches(labels.Set{"environment": environment})).To(BeTrue())
						className := "nginx"

						pathType := networkingv1.PathTypePrefix
						data := &networkingv1.IngressList{
							Items: []networkingv1.Ingress{
								{
									ObjectMeta: metav1.ObjectMeta{
										Name: "ignore-1",
										Labels: map[string]string{
											"environment": environment,
										},
										Annotations: map[string]string{
											"dolittle.io/microservice-id": microserviceID,
										},
									},
									Spec: networkingv1.IngressSpec{
										IngressClassName: &className,
										Rules: []networkingv1.IngressRule{
											{
												Host: "fake",
												IngressRuleValue: networkingv1.IngressRuleValue{
													HTTP: &networkingv1.HTTPIngressRuleValue{
														Paths: []networkingv1.HTTPIngressPath{
															{
																Path:     "/",
																PathType: &pathType,
																Backend:  networkingv1.IngressBackend{},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						}
						return true, data, nil
					})

					items, err := k8sRepo.GetIngressHTTPIngressPath(applicationID, environment, microserviceID)
					Expect(err).To(BeNil())
					Expect(len(items)).To(Equal(1))
					Expect(items[0].Path).To(Equal("/"))
				})

				It("Confirm unique rules", func() {

					clientSet.AddReactor("list", "ingresses", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
						filters := action.(testing.ListActionImpl).ListRestrictions
						Expect(filters.Labels.Matches(labels.Set{"environment": environment})).To(BeTrue())
						className := "nginx"

						pathType := networkingv1.PathTypePrefix
						data := &networkingv1.IngressList{
							Items: []networkingv1.Ingress{
								{
									ObjectMeta: metav1.ObjectMeta{
										Name: "ignore-1",
										Labels: map[string]string{
											"environment": environment,
										},
										Annotations: map[string]string{
											"dolittle.io/microservice-id": microserviceID,
										},
									},
									Spec: networkingv1.IngressSpec{
										IngressClassName: &className,
										Rules: []networkingv1.IngressRule{
											{
												Host: "fake",
												IngressRuleValue: networkingv1.IngressRuleValue{
													HTTP: &networkingv1.HTTPIngressRuleValue{
														Paths: []networkingv1.HTTPIngressPath{
															{
																Path:     "/",
																PathType: &pathType,
																Backend:  networkingv1.IngressBackend{},
															},
														},
													},
												},
											},
											{
												Host: "fake",
												IngressRuleValue: networkingv1.IngressRuleValue{
													HTTP: &networkingv1.HTTPIngressRuleValue{
														Paths: []networkingv1.HTTPIngressPath{
															{
																Path:     "/",
																PathType: &pathType,
																Backend:  networkingv1.IngressBackend{},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						}
						return true, data, nil
					})

					items, err := k8sRepo.GetIngressHTTPIngressPath(applicationID, environment, microserviceID)
					Expect(err).To(BeNil())
					Expect(len(items)).To(Equal(1))
					Expect(items[0].Path).To(Equal("/"))
				})
			})
		})

	})

	When("Getting the microservice name", func() {
		var (
			applicationID  string
			environment    string
			microserviceID string
			want           error
			clientSet      *fake.Clientset
			config         *rest.Config
			k8sRepo        platformK8s.K8sRepo
			logger         *logrus.Logger
		)

		BeforeEach(func() {
			applicationID = "fake-application-123"
			environment = "Dev"
			microserviceID = "fake-microservice-123"
			want = errors.New("fail")
			clientSet = &fake.Clientset{}
			config = &rest.Config{}
			logger, _ = logrusTest.NewNullLogger()
			k8sRepo = platformK8s.NewK8sRepo(clientSet, config, logger.WithField("context", "k8s-repo"))
		})

		It("Failed to talk to kubernetes", func() {
			clientSet.AddReactor("list", "deployments", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, &appsv1.DeploymentList{}, want
			})
			_, err := k8sRepo.GetMicroserviceName(applicationID, environment, microserviceID)
			Expect(err).To(Equal(want))
		})
		It("Not Found", func() {
			clientSet.AddReactor("list", "deployments", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				filters := action.(testing.ListActionImpl).ListRestrictions
				Expect(filters.Labels.Matches(labels.Set{"environment": environment, "microservice": string(selection.Exists)})).To(BeTrue())

				data := &appsv1.DeploymentList{
					Items: []appsv1.Deployment{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "ignore-1",
								Labels: map[string]string{
									"environment": environment,
								},
							},
						},
					},
				}
				return true, data, nil
			})

			_, err := k8sRepo.GetMicroserviceName(applicationID, environment, microserviceID)
			Expect(err).To(Equal(platformK8s.ErrNotFound))
		})

		It("Found", func() {
			clientSet.AddReactor("list", "deployments", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				filters := action.(testing.ListActionImpl).ListRestrictions
				Expect(filters.Labels.Matches(labels.Set{"environment": environment, "microservice": string(selection.Exists)})).To(BeTrue())

				data := &appsv1.DeploymentList{
					Items: []appsv1.Deployment{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "hello-world",
								Labels: map[string]string{
									"environment":  environment,
									"microservice": "hello-world",
								},
								Annotations: map[string]string{
									"dolittle.io/microservice-id": microserviceID,
								},
							},
						},
					},
				}
				return true, data, nil
			})

			name, err := k8sRepo.GetMicroserviceName(applicationID, environment, microserviceID)
			Expect(err).To(BeNil())
			Expect(name).To(Equal("hello-world"))
		})
	})
})
