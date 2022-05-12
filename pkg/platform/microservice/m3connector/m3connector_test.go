package m3connector_test

import (
	"errors"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/dolittle/platform-api/pkg/platform/microservice/m3connector"

	mockM3Connector "github.com/dolittle/platform-api/mocks/pkg/platform/microservice/m3connector"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("M3connector", func() {
	var (
		connector      *m3connector.M3Connector
		mockKafka      *mockM3Connector.KafkaProvider
		mockRepo       *mockM3Connector.K8sRepo
		customer       string
		application    string
		environment    string
		resourcePrefix string
		username       string
		certificate    string
		accessKey      string
		ca             string
	)

	BeforeEach(func() {
		logger, _ := logrusTest.NewNullLogger()
		mockKafka = new(mockM3Connector.KafkaProvider)
		mockRepo = new(mockM3Connector.K8sRepo)
		connector = m3connector.NewM3Connector(mockKafka, mockRepo, logger)
		customer = "8f90ddec-a668-4995-8399-2af9a8731723"
		application = "e0f6a2a4-c136-4a56-80fa-75ff4daa75d6"
		environment = "test"
		resourcePrefix = fmt.Sprintf("cust_%s.app_%s.env_%s.m3connector", customer, application, environment)
		shortCustomer := "8f90ddeca6684995"
		shortApplication := "e0f6a2a4c1364a56"
		username = fmt.Sprintf("%s.%s.%s.m3connector", shortCustomer, shortApplication, environment)
		certificate = "im the certificate"
		accessKey = "im the access key"
		ca = "im the certificate authority"
	})

	var err error
	JustBeforeEach(func() {
		// @joel TODO comment how this works
		mockKafka.On("CreateUser", mock.Anything).Return("", "", nil)
		mockKafka.On("CreateTopic", mock.Anything, mock.Anything).Return(nil)
		mockKafka.On("AddACL", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		mockKafka.On("GetCertificateAuthority").Return("")
		mockRepo.On("UpsertKafkaFiles", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		err = connector.CreateEnvironment(customer, application, environment)
	})

	When("Creating a new environment", func() {
		When("customer is empty", func() {
			BeforeEach(func() {
				customer = ""
			})

			It("should fail", func() {
				Expect(err).ToNot(BeNil())
			})
		})

		XIt("should fail without an application", func() {
			err := connector.CreateEnvironment(customer, "", environment)
			Expect(err).ToNot(BeNil())
		})
		XIt("should fail without an environment", func() {
			err := connector.CreateEnvironment(customer, application, "")
			Expect(err).ToNot(BeNil())
		})

		Describe("with all the necessary data", func() {
			When("user creation fails", func() {
				BeforeEach(func() {
					mockKafka.On(
						"CreateUser",
						username,
					).Return("", "", errors.New("test error"))
				})

				It("should fail", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("everything works", func() {
				It("should not fail", func() {
					Expect(err).To(BeNil())
				})

				It("should create a user with the correct username", func() {
					mockKafka.AssertCalled(GinkgoT(), "CreateUser", username)
				})

				It("should create the 4 topics with correct names and retention", func() {
					changeTopic := fmt.Sprintf("%s.change-events", resourcePrefix)
					inputTopic := fmt.Sprintf("%s.input", resourcePrefix)
					commandTopic := fmt.Sprintf("%s.commands", resourcePrefix)
					receiptsTopic := fmt.Sprintf("%s.command-receipts", resourcePrefix)

					mockKafka.AssertCalled(GinkgoT(), "CreateTopic", commandTopic, int64(-1))
					mockKafka.AssertCalled(GinkgoT(), "CreateTopic", changeTopic, int64(-1))
					mockKafka.AssertCalled(GinkgoT(), "CreateTopic", inputTopic, int64(-1))
					mockKafka.AssertCalled(GinkgoT(), "CreateTopic", receiptsTopic, int64(604800000))
				})

			})

			// mockKafka.On(
			// 	"CreateTopic",
			// 	mock.Anything,
			// 	mock.Anything,
			// ).Return(nil).Times(4)

			// mockKafka.On(
			// 	"AddACL",
			// 	mock.Anything,
			// 	mock.Anything,
			// 	mock.Anything,
			// ).Return(nil).Times(4)

		})

		XIt("should call to create 4 topics with the correct names and retentions", func() {

			changeTopic := fmt.Sprintf("%s.change-events", resourcePrefix)
			inputTopic := fmt.Sprintf("%s.input", resourcePrefix)
			commandTopic := fmt.Sprintf("%s.commands", resourcePrefix)
			receiptsTopic := fmt.Sprintf("%s.command-receipts", resourcePrefix)

			mockKafka.
				On(
					"GetCertificateAuthority",
					mock.Anything,
				).Return(ca).
				On(
					"CreateUser",
					mock.Anything,
				).Return("", "", nil).
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

		XIt("should fail if the topic creation fails", func() {
			changeTopic := fmt.Sprintf("%s.change-events", resourcePrefix)

			mockKafka.
				On(
					"GetCertificateAuthority",
					mock.Anything,
				).Return(ca).
				On(
					"CreateUser",
					mock.Anything,
				).Return("", "", nil).
				On(
					"CreateTopic",
					changeTopic,
					int64(-1),
				).Return(errors.New("topic creation error"))

			err := connector.CreateEnvironment(customer, application, environment)
			Expect(err).ToNot(BeNil())
			mock.AssertExpectationsForObjects(GinkgoT(), mockKafka)
		})

		XIt("should call to create the ACLs for the 4 topics", func() {

			changeTopic := fmt.Sprintf("%s.change-events", resourcePrefix)
			inputTopic := fmt.Sprintf("%s.input", resourcePrefix)
			commandTopic := fmt.Sprintf("%s.commands", resourcePrefix)
			receiptsTopic := fmt.Sprintf("%s.command-receipts", resourcePrefix)
			permission := string(m3connector.ReadWrite)

			mockKafka.
				On(
					"GetCertificateAuthority",
					mock.Anything,
				).Return(ca).
				On(
					"CreateUser",
					mock.Anything,
				).Return("", "", nil)
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

		XIt("should fail if adding an ACL fails", func() {
			changeTopic := fmt.Sprintf("%s.change-events", resourcePrefix)
			permission := string(m3connector.ReadWrite)

			mockKafka.
				On(
					"GetCertificateAuthority",
					mock.Anything,
				).Return(ca).
				On(
					"CreateUser",
					mock.Anything,
				).Return("", "", nil).
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

		XIt("should lowercase the customer, application and environment", func() {
			upperCustomer := strings.ToUpper(customer)
			upperApplication := strings.ToUpper(application)
			upperEnv := strings.ToUpper(environment)

			changeTopic := fmt.Sprintf("%s.change-events", resourcePrefix)
			inputTopic := fmt.Sprintf("%s.input", resourcePrefix)
			commandTopic := fmt.Sprintf("%s.commands", resourcePrefix)
			receiptsTopic := fmt.Sprintf("%s.command-receipts", resourcePrefix)

			mockKafka.
				On(
					"GetCertificateAuthority",
					mock.Anything,
				).Return(ca).
				On(
					"CreateUser",
					mock.Anything,
				).Return("", "", nil).
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
			XIt("should call the config map to be created with the correct values", func() {

				kafkaFiles := m3connector.KafkaFiles{
					AccessKey:            accessKey,
					Certificate:          certificate,
					CertificateAuthority: ca,
				}

				mockKafka.
					On(
						"GetServiceCA",
					).Return(kafkaFiles.CertificateAuthority).
					On(
						"CreateUser",
						mock.Anything,
					).Return(kafkaFiles.Certificate, kafkaFiles.AccessKey, nil).
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

				mockRepo.On(
					"UpsertKafkaFiles",
					application,
					environment,
					kafkaFiles,
				).Return(nil)

				err := connector.CreateEnvironment(customer, application, environment)
				Expect(err).To(BeNil())
				mock.AssertExpectationsForObjects(GinkgoT(), mockKafka)
				mock.AssertExpectationsForObjects(GinkgoT(), mockRepo)
			})
		})
	})
})
