package k8s_test

import (
	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/application/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thoas/go-funk"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
)

var _ = Describe("Setting up an application", func() {
	var (
		customer     dolittleK8s.Tenant
		application  dolittleK8s.Application
		azureGroupId string
		environment  string
	)
	BeforeEach(func() {
		azureGroupId = "azure-fake-123"
		environment = "TODO"
		customer = dolittleK8s.Tenant{
			ID:   "fake-customer-123",
			Name: "TODO",
		}
		application = dolittleK8s.Application{
			ID:   "fake-application-123",
			Name: "TODO",
		}
	})

	When("Creating mongo resource", func() {
		It("Include the application and environment in the name of the file saved", func() {
			settings := k8s.MongoSettings{
				ShareName:       "fake",
				CronJobSchedule: "* * * * *",
				VolumeSize:      "8Gi",
			}
			resource := k8s.NewMongo(environment, customer, application, settings)
			Expect(resource.Service.Name).To(Equal("todo-mongo"))

			expect := `mongodump --host=todo-mongo.application-fake-application-123.svc.cluster.local:27017 --gzip --archive=/mnt/backup/$(APPLICATION)-$(ENVIRONMENT)-$(date +%Y-%m-%d_%H-%M-%S).gz.mongodump`
			Expect(resource.Cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Args[0]).To(Equal(expect))
		})
	})
	When("Creating the environment", func() {
		var (
			resource *networkingv1.NetworkPolicy
		)
		BeforeEach(func() {
			resource = k8s.NewNetworkPolicy(environment, customer, application)
		})
		It("Confirm system api can reach environment", func() {
			found := funk.Contains(resource.Spec.Ingress[0].From, func(policy networkingv1.NetworkPolicyPeer) bool {
				want := k8s.AllowNetworkPolicyForSystemAPI
				return equality.Semantic.DeepDerivative(policy, want)
			})
			Expect(found).To(BeTrue())
		})
		It("Confirm metrics services can reach environment", func() {

			found := funk.Contains(resource.Spec.Ingress[0].From, func(policy networkingv1.NetworkPolicyPeer) bool {
				want := k8s.AllowNetworkPolicyForMonitoring
				return equality.Semantic.DeepDerivative(policy, want)
			})
			Expect(found).To(BeTrue())
		})

		It("Confirm we add the azure ID as a subject of developer", func() {
			rbacResources := k8s.NewDeveloperRole(customer, application, azureGroupId)

			found := funk.Contains(rbacResources.RoleBinding.Subjects, func(subject rbacv1.Subject) bool {
				want := rbacv1.Subject{
					Kind:     "Group",
					APIGroup: "rbac.authorization.k8s.io",
					Name:     azureGroupId,
				}
				return equality.Semantic.DeepDerivative(subject, want)
			})
			Expect(found).To(BeTrue())
		})

		It("Confirm we add the tenantGroup as a subject of developer", func() {
			rbacResources := k8s.NewDeveloperRole(customer, application, azureGroupId)

			found := funk.Contains(rbacResources.RoleBinding.Subjects, func(subject rbacv1.Subject) bool {
				want := rbacv1.Subject{
					Kind:     "Group",
					APIGroup: "rbac.authorization.k8s.io",
					Name:     platform.GetCustomerGroup(customer.ID),
				}
				return equality.Semantic.DeepDerivative(subject, want)
			})
			Expect(found).To(BeTrue())
		})
	})

})
