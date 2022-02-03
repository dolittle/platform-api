package purchaseorderapi_test

import (
	"fmt"

	"github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	. "github.com/dolittle/platform-api/pkg/platform/microservice/purchaseorderapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("For repo", func() {
	var (
		repo Repo

		customer            k8s.Tenant
		application         k8s.Application
		namespace           string
		name                string
		environment         string
		customerTenantID    string
		createInput         platform.HttpInputPurchaseOrderInfo
		existingDeployments []runtime.Object
		customerTenants     []platform.CustomerTenantInfo

		result bool
		err    error
	)

	BeforeEach(func() {
		customer = k8s.Tenant{
			Name: "tenant-name",
			ID:   "67dcf38f-16e4-4b57-bff5-707cff3233ec",
		}
		application = k8s.Application{
			Name: "application-name",
			ID:   "c1e08289-be4b-4557-9457-5de90e0ea54a",
		}
		namespace = fmt.Sprintf("application-%s", application.ID)
		name = "some-name"
		environment = "some-environment"
		customerTenantID = "04b557ed-eb92-476a-b9ef-6c99c1ff9f86"
		customerTenants = []platform.CustomerTenantInfo{
			{
				CustomerTenantID: customerTenantID,
				Ingresses: []platform.CustomerTenantIngress{
					{
						MicroserviceID: "fake-microservice-id",
						Host:           "fake-prefix.fake-host",
						DomainPrefix:   "fake-prefix",
						SecretName:     "fake-prefix",
					},
				},
			},
		}

		createInput = newPurchaseOrderAPICreateInput(customer, application, environment, name)
		existingDeployments = nil
	})
	JustBeforeEach(func() {
		client := fake.NewSimpleClientset(existingDeployments...)
		k8sResourceSpecFactory := NewK8sResourceSpecFactory()
		k8sResource := NewK8sResource(client, k8sResourceSpecFactory)
		repo = NewRepo(k8sResource, k8sResourceSpecFactory, client)
	})

	Describe("when checking if purchase order api exists", func() {
		JustBeforeEach(func() {
			result, err = repo.Exists(namespace, customer, application, customerTenants, createInput)
		})

		Describe("and there is another purchase order api with the same name", func() {
			Describe("in the same namespace and environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							environment,
							name,
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should exist", func() {
					Expect(result).To(BeTrue())
				})
			})
			Describe("in the same namespace but different environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							"other-environment",
							name,
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("in a different namespace but same environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							k8s.Application{
								Name: "another-application",
								ID:   "33607c42-e25b-4982-9a01-e89449db44b2",
							},
							environment,
							name,
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("in a different namespace and environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							k8s.Application{
								Name: "another-application",
								ID:   "33607c42-e25b-4982-9a01-e89449db44b2",
							},
							"other-environment",
							name,
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
		})
		Describe("and there is another purchase order api with a different name", func() {
			Describe("in the same namespace and environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							environment,
							"other-name",
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("in the same namespace but different environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							"other-environment",
							"other-name",
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("in a different namespace but same environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							k8s.Application{
								Name: "another-application",
								ID:   "33607c42-e25b-4982-9a01-e89449db44b2",
							},
							environment,
							"other-name",
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("in a different namespace and environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							k8s.Application{
								Name: "another-application",
								ID:   "33607c42-e25b-4982-9a01-e89449db44b2",
							},
							"other-environment",
							"other-name",
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
		})
		Describe("and there is another microservice kind with the same name", func() {
			Describe("in the same namespace and environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							environment,
							name,
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should exist", func() {
					Expect(result).To(BeTrue())
				})
			})
			Describe("in the same namespace but different environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							"other-environment",
							name,
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("in a different namespace but same environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							k8s.Application{
								Name: "another-application",
								ID:   "33607c42-e25b-4982-9a01-e89449db44b2",
							},
							environment,
							name,
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("in a different namespace and environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							k8s.Application{
								Name: "another-application",
								ID:   "33607c42-e25b-4982-9a01-e89449db44b2",
							},
							"other-environment",
							name,
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
		})
		Describe("and there is another microservice kind with a different name", func() {
			Describe("in the same namespace and environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							environment,
							"other-name",
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("in the same namespace but different environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							"other-environment",
							"other-name",
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("in a different namespace but same environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							k8s.Application{
								Name: "another-application",
								ID:   "33607c42-e25b-4982-9a01-e89449db44b2",
							},
							environment,
							"other-name",
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("in a different namespace and environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							k8s.Application{
								Name: "another-application",
								ID:   "33607c42-e25b-4982-9a01-e89449db44b2",
							},
							"other-environment",
							"other-name",
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
		})
		Describe("and there are multiple deployments in the same namespace and environment", func() {
			Describe("all with other names", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							environment,
							"other-simple-name",
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
						newDeploymentFrom(
							customer,
							application,
							environment,
							"other-name",
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("and a simple with the same name", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							environment,
							"other-name",
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
						newDeploymentFrom(
							customer,
							application,
							environment,
							name,
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should exist", func() {
					Expect(result).To(BeTrue())
				})
			})
			Describe("and a purchase order api with the same name", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							environment,
							"other-name",
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
						newDeploymentFrom(
							customer,
							application,
							environment,
							name,
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should exist", func() {
					Expect(result).To(BeTrue())
				})
			})
		})
		Describe("and there are no deployments", func() {
			It("should not fail", func() {
				Expect(err).To(BeNil())
			})
			It("should not exist", func() {
				Expect(result).ToNot(BeTrue())
			})
		})
	})
	Describe("when checking if environment has purchase order api ", func() {
		JustBeforeEach(func() {
			result, err = repo.EnvironmentHasPurchaseOrderAPI(namespace, createInput)
		})

		Describe("and there is another purchase order api with the same name", func() {
			Describe("in the same namespace and environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							environment,
							name,
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should exist", func() {
					Expect(result).To(BeTrue())
				})
			})
			Describe("in the same namespace but different environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							"other-environment",
							name,
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("in a different namespace but same environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							k8s.Application{
								Name: "another-application",
								ID:   "33607c42-e25b-4982-9a01-e89449db44b2",
							},
							environment,
							name,
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("in a different namespace and environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							k8s.Application{
								Name: "another-application",
								ID:   "33607c42-e25b-4982-9a01-e89449db44b2",
							},
							"other-environment",
							name,
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
		})
		Describe("and there is another purchase order api with a different name", func() {
			Describe("in the same namespace and environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							environment,
							"other-name",
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should exist", func() {
					Expect(result).To(BeTrue())
				})
			})
			Describe("in the same namespace but different environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							"other-environment",
							"other-name",
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("in a different namespace but same environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							k8s.Application{
								Name: "another-application",
								ID:   "33607c42-e25b-4982-9a01-e89449db44b2",
							},
							environment,
							"other-name",
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("in a different namespace and environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							k8s.Application{
								Name: "another-application",
								ID:   "33607c42-e25b-4982-9a01-e89449db44b2",
							},
							"other-environment",
							"other-name",
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
		})
		Describe("and there is another microservice kind with the same name", func() {
			Describe("in the same namespace and environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							environment,
							name,
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("in the same namespace but different environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							"other-environment",
							name,
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("in a different namespace but same environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							k8s.Application{
								Name: "another-application",
								ID:   "33607c42-e25b-4982-9a01-e89449db44b2",
							},
							environment,
							name,
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("in a different namespace and environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							k8s.Application{
								Name: "another-application",
								ID:   "33607c42-e25b-4982-9a01-e89449db44b2",
							},
							"other-environment",
							name,
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
		})
		Describe("and there is another microservice kind with a different name", func() {
			Describe("in the same namespace and environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							environment,
							"other-name",
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("in the same namespace but different environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							"other-environment",
							"other-name",
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("in a different namespace but same environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							k8s.Application{
								Name: "another-application",
								ID:   "33607c42-e25b-4982-9a01-e89449db44b2",
							},
							environment,
							"other-name",
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
			Describe("in a different namespace and environment", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							k8s.Application{
								Name: "another-application",
								ID:   "33607c42-e25b-4982-9a01-e89449db44b2",
							},
							"other-environment",
							"other-name",
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should not exist", func() {
					Expect(result).ToNot(BeTrue())
				})
			})
		})
		Describe("and there are multiple deployments in the same namespace and environment", func() {
			Describe("all with other names", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							environment,
							"other-simple-name",
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
						newDeploymentFrom(
							customer,
							application,
							environment,
							"other-name",
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should exist", func() {
					Expect(result).To(BeTrue())
				})
			})
			Describe("and a simple with the same name", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							environment,
							"other-name",
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
						newDeploymentFrom(
							customer,
							application,
							environment,
							name,
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should exist", func() {
					Expect(result).To(BeTrue())
				})
			})
			Describe("and a purchase order api with the same name", func() {
				BeforeEach(func() {
					existingDeployments = []runtime.Object{
						newDeploymentFrom(
							customer,
							application,
							environment,
							"other-name",
							"d1b21b10-900a-4c23-9bd0-2d49a9d2957c",
							platform.MicroserviceKindSimple,
						),
						newDeploymentFrom(
							customer,
							application,
							environment,
							name,
							"ef97a13b-2597-42a3-9fcb-161add2264c7",
							platform.MicroserviceKindPurchaseOrderAPI,
						),
					}
				})
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})
				It("should exist", func() {
					Expect(result).To(BeTrue())
				})
			})
		})
		Describe("and there are no deployments", func() {
			It("should not fail", func() {
				Expect(err).To(BeNil())
			})
			It("should not exist", func() {
				Expect(result).ToNot(BeTrue())
			})
		})
	})
})

func newPurchaseOrderAPICreateInput(customer k8s.Tenant, application k8s.Application, environment, name string) platform.HttpInputPurchaseOrderInfo {
	return platform.HttpInputPurchaseOrderInfo{
		MicroserviceBase: platform.MicroserviceBase{
			Dolittle: platform.HttpInputDolittle{
				TenantID:       customer.ID,
				ApplicationID:  application.ID,
				MicroserviceID: "ef97a13b-2597-42a3-9fcb-161add2264c7",
			},
			Name:        name,
			Kind:        platform.MicroserviceKindPurchaseOrderAPI,
			Environment: environment,
		},
		Extra: platform.HttpInputPurchaseOrderExtra{
			RawDataLogName: "raw-data-log-name",
			Headimage:      "head-image",
			Runtimeimage:   "runtime-image",
			Webhooks:       []platform.RawDataLogIngestorWebhookConfig{},
		},
	}
}

func newDeploymentFrom(customer k8s.Tenant, application k8s.Application, environment, name, id string, kind platform.MicroserviceKind) *v1.Deployment {
	return k8s.NewDeployment(k8s.Microservice{
		ID:          id,
		Name:        name,
		Tenant:      customer,
		Application: application,
		Environment: environment,
		Kind:        kind,
	}, "head-image:shouldnt-matter", "runtime-image:shouldnt-matter")
}
