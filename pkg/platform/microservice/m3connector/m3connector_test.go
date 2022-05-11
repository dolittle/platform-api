package m3connector_test

import (
	"errors"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/testing"

	"github.com/dolittle/platform-api/pkg/platform/microservice/m3connector"

	mockM3Connector "github.com/dolittle/platform-api/mocks/pkg/platform/microservice/m3connector"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("M3connector", func() {
	var (
		connector      *m3connector.M3Connector
		mockKafka      *mockM3Connector.KafkaProvider
		customer       string
		application    string
		environment    string
		resourcePrefix string
		username       string
		clientSet      *fake.Clientset
	)
	BeforeEach(func() {
		logger, _ := logrusTest.NewNullLogger()
		mockKafka = new(mockM3Connector.KafkaProvider)
		connector = m3connector.NewM3Connector(mockKafka, logger)
		customer = "8f90ddec-a668-4995-8399-2af9a8731723"
		application = "e0f6a2a4-c136-4a56-80fa-75ff4daa75d6"
		environment = "test"
		resourcePrefix = fmt.Sprintf("cust_%s.app_%s.env_%s.m3connector", customer, application, environment)
		shortCustomer := "8f90ddeca6684995"
		shortApplication := "e0f6a2a4c1364a56"
		username = fmt.Sprintf("%s.%s.%s.m3connector", shortCustomer, shortApplication, environment)
	})
	Describe("Creating a new environment", func() {
		It("should fail without a customer", func() {
			err := connector.CreateEnvironment("", application, environment)
			Expect(err).ToNot(BeNil())
		})
		It("should fail without an application", func() {
			err := connector.CreateEnvironment(customer, "", environment)
			Expect(err).ToNot(BeNil())
		})
		It("should fail without an environment", func() {
			err := connector.CreateEnvironment(customer, application, "")
			Expect(err).ToNot(BeNil())
		})

		It("should call to create an user with the correct username", func() {
			mockKafka.On(
				"CreateUser",
				username,
			).Return(nil)

			mockKafka.On(
				"CreateTopic",
				mock.Anything,
				mock.Anything,
			).Return(nil).Times(4)

			mockKafka.On(
				"AddACL",
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(nil).Times(4)

			err := connector.CreateEnvironment(customer, application, environment)
			Expect(err).To(BeNil())
			mock.AssertExpectationsForObjects(GinkgoT(), mockKafka)
		})

		It("should fail if the user creation fails", func() {
			mockKafka.On(
				"CreateUser",
				username,
			).Return(errors.New("test error"))

			err := connector.CreateEnvironment(customer, application, environment)
			Expect(err).ToNot(BeNil())
			mock.AssertExpectationsForObjects(GinkgoT(), mockKafka)
		})

		It("should call to create 4 topics with the correct names and retentions", func() {

			changeTopic := fmt.Sprintf("%s.change-events", resourcePrefix)
			inputTopic := fmt.Sprintf("%s.input", resourcePrefix)
			commandTopic := fmt.Sprintf("%s.commands", resourcePrefix)
			receiptsTopic := fmt.Sprintf("%s.command-receipts", resourcePrefix)

			mockKafka.
				On(
					"CreateUser",
					mock.Anything,
				).Return(nil).
				On(
					"CreateTopic",
					changeTopic,
					int64(-1),
				).Return(nil).
				On(
					"CreateTopic",
					inputTopic,
					int64(-1),
				).Return(nil).
				On(
					"CreateTopic",
					commandTopic,
					int64(-1),
				).Return(nil).
				On(
					"CreateTopic",
					receiptsTopic,
					int64(604800000),
				).Return(nil)

			mockKafka.On(
				"AddACL",
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(nil).Times(4)

			err := connector.CreateEnvironment(customer, application, environment)
			Expect(err).To(BeNil())
			mock.AssertExpectationsForObjects(GinkgoT(), mockKafka)
		})

		It("should fail if the topic creation fails", func() {
			changeTopic := fmt.Sprintf("%s.change-events", resourcePrefix)

			mockKafka.
				On(
					"CreateUser",
					mock.Anything,
				).Return(nil).
				On(
					"CreateTopic",
					changeTopic,
					int64(-1),
				).Return(errors.New("topic creation error"))

			err := connector.CreateEnvironment(customer, application, environment)
			Expect(err).ToNot(BeNil())
			mock.AssertExpectationsForObjects(GinkgoT(), mockKafka)
		})

		It("should call to create the ACLs for the 4 topics", func() {

			changeTopic := fmt.Sprintf("%s.change-events", resourcePrefix)
			inputTopic := fmt.Sprintf("%s.input", resourcePrefix)
			commandTopic := fmt.Sprintf("%s.commands", resourcePrefix)
			receiptsTopic := fmt.Sprintf("%s.command-receipts", resourcePrefix)
			permission := string(m3connector.ReadWrite)

			mockKafka.
				On(
					"CreateUser",
					mock.Anything,
				).Return(nil)
			mockKafka.On(
				"CreateTopic",
				mock.Anything,
				mock.Anything,
			).Return(nil).Times(4)

			mockKafka.
				On(
					"AddACL",
					changeTopic,
					username,
					permission,
				).Return(nil).
				On(
					"AddACL",
					inputTopic,
					username,
					permission,
				).Return(nil).
				On(
					"AddACL",
					commandTopic,
					username,
					permission,
				).Return(nil).
				On(
					"AddACL",
					receiptsTopic,
					username,
					permission,
				).Return(nil)

			err := connector.CreateEnvironment(customer, application, environment)
			Expect(err).To(BeNil())
			mock.AssertExpectationsForObjects(GinkgoT(), mockKafka)
		})

		It("should fail if adding an ACL fails", func() {
			changeTopic := fmt.Sprintf("%s.change-events", resourcePrefix)
			permission := string(m3connector.ReadWrite)

			mockKafka.
				On(
					"CreateUser",
					mock.Anything,
				).Return(nil).
				On(
					"CreateTopic",
					changeTopic,
					int64(-1),
				).Return(nil).
				On(
					"AddACL",
					changeTopic,
					username,
					permission,
				).Return(errors.New("error adding acl"))

			err := connector.CreateEnvironment(customer, application, environment)
			Expect(err).ToNot(BeNil())
			mock.AssertExpectationsForObjects(GinkgoT(), mockKafka)
		})

		It("should lowercase the customer, application and environment", func() {
			upperCustomer := strings.ToUpper(customer)
			upperApplication := strings.ToUpper(application)
			upperEnv := strings.ToUpper(environment)

			changeTopic := fmt.Sprintf("%s.change-events", resourcePrefix)
			inputTopic := fmt.Sprintf("%s.input", resourcePrefix)
			commandTopic := fmt.Sprintf("%s.commands", resourcePrefix)
			receiptsTopic := fmt.Sprintf("%s.command-receipts", resourcePrefix)

			mockKafka.On(
				"CreateUser",
				username,
			).Return(nil)

			mockKafka.
				On(
					"CreateUser",
					mock.Anything,
				).Return(nil).
				On(
					"CreateTopic",
					changeTopic,
					mock.Anything,
				).Return(nil).
				On(
					"CreateTopic",
					inputTopic,
					mock.Anything,
				).Return(nil).
				On(
					"CreateTopic",
					commandTopic,
					mock.Anything,
				).Return(nil).
				On(
					"CreateTopic",
					receiptsTopic,
					mock.Anything,
				).Return(nil)

			mockKafka.
				On(
					"AddACL",
					changeTopic,
					username,
					mock.Anything,
				).Return(nil).
				On(
					"AddACL",
					inputTopic,
					username,
					mock.Anything,
				).Return(nil).
				On(
					"AddACL",
					commandTopic,
					username,
					mock.Anything,
				).Return(nil).
				On(
					"AddACL",
					receiptsTopic,
					username,
					mock.Anything,
				).Return(nil)

			err := connector.CreateEnvironment(upperCustomer, upperApplication, upperEnv)
			Expect(err).To(BeNil())
			mock.AssertExpectationsForObjects(GinkgoT(), mockKafka)
		})

		When("writing the credentials and config to the kafka-files configmap", func() {
			It("should create the confimap if it doesn't exist", func() {

				mockKafka.
					On(
						"CreateUser",
						mock.Anything,
					).Return(nil).
					On(
						"CreateTopic",
						mock.Anything,
						mock.Anything,
					).Return(nil).
					On(
						"AddACL",
						mock.Anything,
						mock.Anything,
						mock.Anything,
					).Return(nil)

				// make the get return no configmap, simulating a missing configmap
				clientSet.AddReactor("get", "configmaps", func(action testing.Action) (bool, runtime.Object, error) {
					return true, nil, nil
				})
				clientSet.AddReactor("create", "configmaps", func(action testing.Action) (bool, runtime.Object, error) {
					createAction := action.(testing.CreateAction)
					originalObj := createAction.GetObject()
					configMap := originalObj.(*corev1.ConfigMap)

					Expect(configMap.Data["accessKey.pem"]).To(Equal())

					return true, nil, nil
				})

				err := connector.CreateEnvironment(customer, application, environment)
				Expect(err).To(BeNil())
				mock.AssertExpectationsForObjects(GinkgoT(), mockKafka)
			})

		})

	})
})
