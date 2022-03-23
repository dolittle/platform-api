package k8s_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
	"github.com/thoas/go-funk"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/testing"

	"github.com/dolittle/platform-api/pkg/k8s"
	applicationK8s "github.com/dolittle/platform-api/pkg/platform/application/k8s"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	platformApplication "github.com/dolittle/platform-api/pkg/platform/application"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	k8sSimple "github.com/dolittle/platform-api/pkg/platform/microservice/simple/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/welcome"
	"github.com/dolittle/platform-api/pkg/platform/storage"

	mockStorage "github.com/dolittle/platform-api/mocks/pkg/platform/storage"

	v1beta1 "k8s.io/api/batch/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
)

var _ = Describe("Repo", func() {
	var (
		clientSet           *fake.Clientset
		config              *rest.Config
		k8sRepo             platformK8s.K8sRepo
		gitRepo             *mockStorage.Repo
		logger              *logrus.Logger
		platformEnvironment string

		terraformCustomer     platform.TerraformCustomer
		terraformApplication  platform.TerraformApplication
		application           storage.JSONApplication
		welcomeMicroserviceID string
		environment           string
	)

	BeforeEach(func() {
		clientSet = fake.NewSimpleClientset()
		config = &rest.Config{}
		logger, _ = logrusTest.NewNullLogger()
		k8sRepo = platformK8s.NewK8sRepo(clientSet, config, logger)
		gitRepo = new(mockStorage.Repo)
		platformEnvironment = "dev"

		terraformCustomer = platform.TerraformCustomer{
			Name:                      "test-customer",
			GUID:                      "fake-customer-id",
			AzureStorageAccountName:   "fake-azure-storage-acccount-name",
			AzureStorageAccountKey:    "fake-azure-storage-acccount-key",
			ContainerRegistryName:     "fake-container-registry-name",
			ContainerRegistryPassword: "fake-container-registry-password",
			ContainerRegistryUsername: "fake-container-registry-username",
			ModuleName:                "fake-module-name",
			PlatformEnvironment:       platformEnvironment,
		}

		terraformApplication = platform.TerraformApplication{
			Customer:      terraformCustomer,
			GroupID:       "fake-azure-group-id",
			GUID:          "fake-application-id",
			ApplicationID: "fake-application-id",
			Kind:          "dolittle-application-with-resources",
			Name:          "test-application",
		}
		isProduction := false
		welcomeImage := welcome.Image
		logContext := logger
		welcomeMicroserviceID = "fake-microservice-id"
		environment = "Dev"
		application = storage.JSONApplication{
			ID:           terraformApplication.ApplicationID,
			Name:         terraformApplication.Name,
			CustomerID:   terraformCustomer.GUID,
			CustomerName: terraformCustomer.Name,
			Environments: []storage.JSONEnvironment{
				{
					Name: environment,
					CustomerTenants: []platform.CustomerTenantInfo{
						dolittleK8s.NewDevelopmentCustomerTenantInfo(environment, welcomeMicroserviceID),
					},
					WelcomeMicroserviceID: welcomeMicroserviceID,
				},
			},
		}

		k8sDolittleRepo := platformK8s.NewK8sRepo(clientSet, config, logContext.WithField("context", "k8s-repo"))
		k8sRepoV2 := k8s.NewRepo(clientSet, logContext.WithField("context", "k8s-repo-v2"))
		simpleRepo := k8sSimple.NewSimpleRepo(clientSet, k8sDolittleRepo, k8sRepoV2, isProduction)
		// TODO refactor when it works

		gitRepo.On(
			"SaveMicroservice",
			application.CustomerID,
			application.ID,
			environment,
			welcomeMicroserviceID,
			mock.Anything,
		).Return(nil)
		err := platformApplication.CreateApplicationAndEnvironmentAndWelcomeMicroservice(
			clientSet,
			gitRepo,
			simpleRepo,
			k8sRepo,
			application,
			terraformCustomer,
			terraformApplication,
			isProduction,
			welcomeImage,
			logContext,
		)

		Expect(err).To(BeNil())
	})

	It("Expected output", func() {
		type test struct {
			Kind   string
			Name   string
			Expect func(ret runtime.Object)
		}

		tests := []test{
			{
				Kind: "Namespace",
				Name: "application-fake-application-id",
				Expect: func(ret runtime.Object) {
					namespace := ret.(*corev1.Namespace)
					Expect(namespace.Name).To(Equal(platformK8s.GetApplicationNamespace(application.ID)))
				},
			},
			{
				Kind: "Secret",
				Name: "acr",
				Expect: func(ret runtime.Object) {
					secret := ret.(*corev1.Secret)
					Expect(secret.Name).To(Equal("acr"))

					data := secret.StringData[".dockerconfigjson"]
					var config applicationK8s.DockerConfigJSON
					json.Unmarshal([]byte(data), &config)
					Expect(config.Auths["fake-container-registry-name.azurecr.io"].Username).To(Equal(terraformCustomer.ContainerRegistryUsername))
					Expect(config.Auths["fake-container-registry-name.azurecr.io"].Password).To(Equal(terraformCustomer.ContainerRegistryPassword))
				},
			},
			{
				Kind: "Role",
				Name: "developer",
				Expect: func(ret runtime.Object) {
					role := ret.(*rbacv1.Role)
					Expect(role.Name).To(Equal("developer"))
				},
			},
			{
				Kind: "RoleBinding",
				Name: "developer",
				Expect: func(ret runtime.Object) {
					roleBinding := ret.(*rbacv1.RoleBinding)
					Expect(roleBinding.Name).To(Equal("developer"))

					found := funk.Contains(roleBinding.Subjects, func(subject rbacv1.Subject) bool {
						want := rbacv1.Subject{
							Kind:     "Group",
							APIGroup: "rbac.authorization.k8s.io",
							Name:     terraformApplication.GroupID,
						}
						return equality.Semantic.DeepDerivative(subject, want)
					})
					Expect(found).To(BeTrue(), "Confirm azure grouid has access")

					found = funk.Contains(roleBinding.Subjects, func(subject rbacv1.Subject) bool {
						want := rbacv1.Subject{
							Kind:     "Group",
							APIGroup: "rbac.authorization.k8s.io",
							Name:     platform.GetCustomerGroup(application.CustomerID),
						}
						return equality.Semantic.DeepDerivative(subject, want)
					})
					Expect(found).To(BeTrue(), "Confirm customer group has access")
				},
			},
			{
				Kind: "ServiceAccount",
				Name: "devops",
				Expect: func(ret runtime.Object) {
					serviceAccount := ret.(*corev1.ServiceAccount)
					Expect(serviceAccount.Name).To(Equal("devops"))
				},
			},
			{
				Kind: "RoleBinding",
				Name: "devops",
				Expect: func(ret runtime.Object) {
					// TODO why twice?
					roleBinding := ret.(*rbacv1.RoleBinding)
					Expect(roleBinding.Name).To(Equal("devops"))
				},
			},
			{
				Kind: "RoleBinding",
				Name: "devops",
				Expect: func(ret runtime.Object) {
					roleBinding := ret.(*rbacv1.RoleBinding)
					Expect(roleBinding.Name).To(Equal("devops"))
				},
			},
			{
				Kind: "Secret",
				Name: "storage-account-secret",
				Expect: func(ret runtime.Object) {
					secret := ret.(*corev1.Secret)
					Expect(secret.Name).To(Equal("storage-account-secret"))
					data := secret.StringData

					Expect(data["azurestorageaccountname"]).To(Equal(terraformCustomer.AzureStorageAccountName))
					Expect(data["azurestorageaccountkey"]).To(Equal(terraformCustomer.AzureStorageAccountKey))
				},
			},
			{
				Kind: "ConfigMap",
				Name: "dev-tenants",
				Expect: func(ret runtime.Object) {
					configMap := ret.(*corev1.ConfigMap)
					Expect(configMap.Name).To(Equal("dev-tenants"))
					var tenants platform.RuntimeTenantsIDS
					json.Unmarshal([]byte(configMap.Data["tenants.json"]), &tenants)
					want := application.Environments[0].CustomerTenants[0].CustomerTenantID
					_, ok := tenants[want]
					Expect(ok).To(BeTrue())
					b, _ := json.Marshal(tenants[want])
					Expect(string(b)).To(Equal(`{}`))
				},
			},
			{
				Kind: "NetworkPolicy",
				Name: "dev",
				Expect: func(ret runtime.Object) {
					resource := ret.(*networkingv1.NetworkPolicy)
					Expect(resource.Name).To(Equal("dev"))

					found := funk.Contains(resource.Spec.Ingress[0].From, func(policy networkingv1.NetworkPolicyPeer) bool {
						want := applicationK8s.AllowNetworkPolicyForSystemAPI
						return equality.Semantic.DeepDerivative(policy, want)
					})
					Expect(found).To(BeTrue(), "Access for system-api")

					found = funk.Contains(resource.Spec.Ingress[0].From, func(policy networkingv1.NetworkPolicyPeer) bool {
						want := applicationK8s.AllowNetworkPolicyForMonitoring
						return equality.Semantic.DeepDerivative(policy, want)
					})
					Expect(found).To(BeTrue(), "Access for monitoring")
				},
			},
			{
				Kind: "Service",
				Name: "dev-mongo",
				Expect: func(ret runtime.Object) {
					service := ret.(*corev1.Service)
					Expect(service.Name).To(Equal("dev-mongo"))
				},
			},
			{
				Kind: "StatefulSet",
				Name: "dev-mongo",
				Expect: func(ret runtime.Object) {
					statefulSet := ret.(*appsv1.StatefulSet)
					Expect(statefulSet.Name).To(Equal("dev-mongo"))
				},
			},
			{
				Kind: "CronJob",
				Name: "dev-mongo-backup",
				Expect: func(ret runtime.Object) {
					cronJob := ret.(*v1beta1.CronJob)
					Expect(cronJob.Name).To(Equal("dev-mongo-backup"))
				},
			},
			{
				Kind: "Role",
				Name: "developer",
				Expect: func(ret runtime.Object) {
					role := ret.(*rbacv1.Role)
					Expect(role.Name).To(Equal("developer"))
				},
			},
			// TODO do I want to see this?
			{
				Kind: "RoleBinding",
				Name: "developer",
				Expect: func(ret runtime.Object) {
					roleBinding := ret.(*rbacv1.RoleBinding)
					Expect(roleBinding.Name).To(Equal("developer-local-dev"))
				},
			},
			{
				Kind: "ConfigMap",
				Name: "dev-welcome-dolittle",
				Expect: func(ret runtime.Object) {
					configMap := ret.(*corev1.ConfigMap)
					Expect(configMap.Name).To(Equal("dev-welcome-dolittle"))
				},
			},
			{
				Kind: "ConfigMap",
				Name: "dev-welcome-env-variables",
				Expect: func(ret runtime.Object) {
					configMap := ret.(*corev1.ConfigMap)
					Expect(configMap.Name).To(Equal("dev-welcome-env-variables"))
				},
			},
			{
				Kind: "ConfigMap",
				Name: "dev-welcome-config-files",
				Expect: func(ret runtime.Object) {
					configMap := ret.(*corev1.ConfigMap)
					Expect(configMap.Name).To(Equal("dev-welcome-config-files"))
				},
			},
			{
				Kind: "Secret",
				Name: "dev-welcome-secret-env-variables",
				Expect: func(ret runtime.Object) {
					secret := ret.(*corev1.Secret)
					Expect(secret.Name).To(Equal("dev-welcome-secret-env-variables"))
				},
			},
			{
				Kind: "Service",
				Name: "dev-welcome",
				Expect: func(ret runtime.Object) {
					service := ret.(*corev1.Service)
					Expect(service.Name).To(Equal("dev-welcome"))
				},
			},
			{
				Kind: "Deployment",
				Name: "dev-welcome",
				Expect: func(ret runtime.Object) {
					deployment := ret.(*appsv1.Deployment)
					Expect(deployment.Name).To(Equal("dev-welcome"))
				},
			},
			{
				Kind: "Role",
				Name: "developer",
				Expect: func(ret runtime.Object) {
					role := ret.(*rbacv1.Role)
					Expect(role.Name).To(Equal("developer"))
				},
			},
			{
				Kind: "Role",
				Name: "developer",
				Expect: func(ret runtime.Object) {
					role := ret.(*rbacv1.Role)
					Expect(role.Name).To(Equal("developer"))
				},
			},
			{
				Kind: "Ingress",
				Name: "dev-welcome-445f8ea", // THIS MIGHT CHANGE
				Expect: func(ret runtime.Object) {
					ingress := ret.(*networkingv1.Ingress)
					Expect(ingress.Name).To(Equal("dev-welcome-445f8ea"))
				},
			},
			{
				Kind: "Ingress",
				Name: "dev-welcome-ingress",
				Expect: func(ret runtime.Object) {
					networkPolicy := ret.(*networkingv1.NetworkPolicy)
					Expect(networkPolicy.Name).To(Equal("dev-welcome-ingress"))
				},
			},
		}

		// TODO bring back
		createActions := getCreateActions(clientSet)

		for index, test := range tests {
			ret := createActions[index].GetObject()
			Expect(ret).ToNot(BeNil())
			test.Expect(createActions[index].GetObject())
		}
		Expect(len(createActions)).To(Equal(len(tests)))
	})
})

func getCreateActions(clientSet *fake.Clientset) []testing.CreateAction {
	var actions []testing.CreateAction
	for _, action := range clientSet.Actions() {
		if create, ok := action.(testing.CreateAction); ok {
			actions = append(actions, create)
		}
	}
	return actions
}
