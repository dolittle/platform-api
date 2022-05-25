package k8s_test

import (
	"encoding/json"
	"fmt"

	"github.com/dolittle/platform-api/pkg/platform/microservice/m3connector"
	"github.com/dolittle/platform-api/pkg/platform/microservice/m3connector/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"

	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/testing"
)

var _ = Describe("Repo", func() {
	var (
		clientSet        *fake.Clientset
		repo             m3connector.K8sRepo
		logger           *logrus.Logger
		err              error
		applicationID    string
		environment      string
		kafkaFiles       m3connector.KafkaFiles
		updatedConfigMap *corev1.ConfigMap
		createdConfigMap *corev1.ConfigMap
		updatedConfig    m3connector.KafkaConfig
		getError         error
		getConfigMap     *corev1.ConfigMap
		getNamespace     *corev1.Namespace
	)

	BeforeEach(func() {

		clientSet = &fake.Clientset{}
		logger, _ = logrusTest.NewNullLogger()
		repo = k8s.NewM3ConnectorRepo(clientSet, false, logger)
		applicationID = "fb8836a0-4fc4-437d-8f25-bb200662f327"
		environment = "test"
		kafkaFiles = m3connector.KafkaFiles{
			AccessKey:            "test-key",
			CertificateAuthority: "test-authority",
			Certificate:          "test-certificate",
			Config: m3connector.KafkaConfig{
				BrokerUrl: "test-url",
				Topics: []string{
					"test",
					"topic",
					"s",
				},
			},
		}
		updatedConfig = m3connector.KafkaConfig{}
		getConfigMap = nil
		getError = nil
		getNamespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("application-%s", applicationID),
				Labels: map[string]string{
					"application": "TestApp",
					"tenant":      "TestCustomer",
				},
				Annotations: map[string]string{
					"dolittle.io/application-id": applicationID,
					"dolittle.io/tenant-id":      "53816a9b-cc60-4889-8a2f-5a5d51526371",
				},
			},
		}
	})

	JustBeforeEach(func() {
		clientSet.AddReactor("get", "configmaps", func(action testing.Action) (bool, runtime.Object, error) {
			return true, getConfigMap, getError
		})

		clientSet.AddReactor("update", "configmaps", func(action testing.Action) (bool, runtime.Object, error) {
			updateAction := action.(testing.UpdateAction)
			updatedConfigMap = updateAction.GetObject().(*corev1.ConfigMap)
			json.Unmarshal([]byte(updatedConfigMap.Data["config.json"]), &updatedConfig)
			return true, updatedConfigMap, nil
		})

		clientSet.AddReactor("create", "configmaps", func(action testing.Action) (bool, runtime.Object, error) {
			createAction := action.(testing.CreateAction)
			createdConfigMap = createAction.GetObject().(*corev1.ConfigMap)
			return true, createdConfigMap, nil
		})
		clientSet.AddReactor("get", "namespaces", func(action testing.Action) (bool, runtime.Object, error) {
			return true, getNamespace, nil
		})

		err = repo.UpsertKafkaFiles(applicationID, environment, kafkaFiles)
	})

	When("the kafka files configmap didn't already exist", func() {
		BeforeEach(func() {
			getError = k8sErrors.NewNotFound(schema.ParseGroupResource("corev1.configmaps"), fmt.Sprintf("%s-kafka-files", environment))
		})

		When("everything works", func() {
			It("should not fail", func() {
				Expect(err).To(BeNil())
			})

			It("should create a new configmap", func() {
				Expect(createdConfigMap).ToNot(BeNil())
			})
			It("should write the access key", func() {
				Expect(createdConfigMap.Data["accessKey.pem"]).To(Equal(kafkaFiles.AccessKey))
			})
			It("should write the certificate", func() {
				Expect(createdConfigMap.Data["certificate.pem"]).To(Equal(kafkaFiles.Certificate))
			})
			It("should write the access key", func() {
				Expect(createdConfigMap.Data["ca.pem"]).To(Equal(kafkaFiles.CertificateAuthority))
			})
			It("should have all the correct labels", func() {
				Expect(createdConfigMap.Labels["application"]).To(Equal(getNamespace.Labels["application"]))
				Expect(createdConfigMap.Labels["tenant"]).To(Equal(getNamespace.Labels["tenant"]))
				Expect(createdConfigMap.Labels["environment"]).To(Equal("Test"))
			})
			It("should have all the correct annotations", func() {
				Expect(createdConfigMap.Annotations["dolittle.io/application-id"]).To(Equal(getNamespace.Annotations["dolittle.io/application-id"]))
				Expect(createdConfigMap.Annotations["dolittle.io/tenant-id"]).To(Equal(getNamespace.Annotations["dolittle.io/tenant-id"]))
			})
		})
	})

	When("Updating an already existing kafka file", func() {

		BeforeEach(func() {
			getConfigMap = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-kafka-files", environment),
				},
				Data: map[string]string{},
			}
		})

		Describe("with all the necessary data", func() {
			When("everything works", func() {
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})

				It("should overwrite the config.json with the given config", func() {
					Expect(updatedConfig).To(Equal(kafkaFiles.Config))
				})

				It("should overwrite the access key", func() {
					Expect(updatedConfigMap.Data["accessKey.pem"]).To(Equal(kafkaFiles.AccessKey))
				})
				It("should overwrite the certificate", func() {
					Expect(updatedConfigMap.Data["certificate.pem"]).To(Equal(kafkaFiles.Certificate))
				})
				It("should overwrite the access key", func() {
					Expect(updatedConfigMap.Data["ca.pem"]).To(Equal(kafkaFiles.CertificateAuthority))
				})
			})
		})
	})
})
