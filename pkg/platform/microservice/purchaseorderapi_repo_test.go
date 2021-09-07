package microservice_test

import (
	. "github.com/dolittle-entropy/platform-api/pkg/platform/microservice"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/rawdatalog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var _ = Describe("PurchaseorderapiRepo", func() {
	var (
		client         *kubernetes.Clientset
		config         *rest.Config
		k8sRepo        platform.K8sRepo
		rawDataLogRepo rawdatalog.RawDataLogIngestorRepo
		repo           PurchaseOrderAPIRepo
	)

	BeforeEach(func() {
		client = nil
		config = nil
		k8sRepo = platform.NewK8sRepo(client, config)
		rawDataLogRepo = rawdatalog.NewRawDataLogIngestorRepo(k8sRepo, client)
		repo = NewPurchaseOrderAPIRepo(client, rawDataLogRepo)
	})

	Describe("when modifying environmental variables", func() {
		var (
			configMap    *corev1.ConfigMap
			microservice k8s.Microservice
		)

		BeforeEach(func() {
			configMap = &corev1.ConfigMap{}
			microservice = k8s.Microservice{
				Environment: "Prod",
				Name:        "ramanujan",
				Application: k8s.Application{
					Name: "einstein",
					ID:   "12345-789",
				},
			}

			repo.ModifyEnvironmentVariablesConfigMap(configMap, microservice)
		})

		It("should set LOG_LEVEL to 'debug'", func() {
			Expect(configMap.Data["LOG_LEVEL"]).To(Equal("debug"))
		})
		It("should set the correct DATABASE_READMODELS_URL", func() {
			Expect(configMap.Data["DATABASE_READMODELS_URL"]).To(Equal("mongodb://prod-mongo.application-12345-789.svc.cluster.local:27017"))
		})
		It("should set the correct DATABASE_READMODELS_NAME", func() {
			Expect(configMap.Data["DATABASE_READMODELS_NAME"]).To(Equal("einstein_prod_ramanujan_readmodels"))
		})
		It("should set NODE_ENV to 'production'", func() {
			Expect(configMap.Data["NODE_ENV"]).To(Equal("production"))
		})
		It("should set TENANT to the todo-customer-tenant-id", func() {
			Expect(configMap.Data["TENANT"]).To(Equal(TodoCustomersTenantID))
		})
		It("should set SERVER_PORT to '8080'", func() {
			Expect(configMap.Data["SERVER_PORT"]).To(Equal("8080"))
		})
		It("should set the correct NATS_CLUSTER_URL", func() {
			Expect(configMap.Data["NATS_CLUSTER_URL"]).To(Equal("prod-rawdatalogv1-nats.application-12345-789.svc.cluster.local:4222"))
		})
		It("should set NATS_START_FROM_BEGINNING to 'false'", func() {
			Expect(configMap.Data["NATS_START_FROM_BEGINNING"]).To(Equal("false"))
		})
		It("should set LOG_OUTPUT_FORMAT to 'json'", func() {
			Expect(configMap.Data["LOG_OUTPUT_FORMAT"]).To(Equal("json"))
		})
	})
})
