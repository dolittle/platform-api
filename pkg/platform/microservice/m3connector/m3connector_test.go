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
		err            error
		changeTopic    string
		inputTopic     string
		commandTopic   string
		receiptsTopic  string
		permission     string
		kafkaFiles     m3connector.KafkaFiles
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

		kafkaFiles = m3connector.KafkaFiles{
			AccessKey:            accessKey,
			Certificate:          certificate,
			CertificateAuthority: ca,
		}

		changeTopic = fmt.Sprintf("%s.change-events", resourcePrefix)
		inputTopic = fmt.Sprintf("%s.input", resourcePrefix)
		commandTopic = fmt.Sprintf("%s.commands", resourcePrefix)
		receiptsTopic = fmt.Sprintf("%s.command-receipts", resourcePrefix)

		permission = string(m3connector.ReadWrite)
	})

	// this section is run after all BeforeEach() blocks but before the It() blocks, meaning we can setup some default
	// behaviours for the mocks without having to always define the exact behaviour. This works by allowing the On()
	// setups on each mock in their BeforeEach()'s to be added to the mocks ExpectedCalls slice before the default
	// behaviour so that when the mock tries to check for the calls it will find the specific setups first in the slice
	// before the default behaviour.
	JustBeforeEach(func() {
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

		When("application is empty", func() {
			BeforeEach(func() {
				application = ""
			})

			It("should fail", func() {
				Expect(err).ToNot(BeNil())
			})
		})

		When("environment is empty", func() {
			BeforeEach(func() {
				environment = ""
			})

			It("should fail", func() {
				Expect(err).ToNot(BeNil())
			})
		})

		Describe("with all the necessary data", func() {
			When("user creation fails", func() {
				BeforeEach(func() {
					mockKafka.On(
						"CreateUser",
						username,
					).Return("", "", errors.New("user error"))
				})

				It("should fail", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("topic creation fails", func() {
				BeforeEach(func() {
					mockKafka.On(
						"CreateTopic",
						mock.Anything,
						mock.Anything,
					).Return(errors.New("topic error"))
				})

				It("should fail", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("topic creation fails", func() {
				BeforeEach(func() {
					mockKafka.On(
						"AddACL",
						mock.Anything,
						mock.Anything,
						mock.Anything,
					).Return(errors.New("ACL error"))
				})

				It("should fail", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("writing the kafka-files fails", func() {
				BeforeEach(func() {
					mockRepo.On(
						"UpsertKafkaFiles",
						mock.Anything,
						mock.Anything,
						mock.Anything,
					).Return(errors.New("kafka-files error"))
				})

				It("should fail", func() {
					Expect(err).ToNot(BeNil())
				})
			})

			When("everything works", func() {
				BeforeEach(func() {
					mockKafka.
						On(
							"GetCertificateAuthority",
						).Return(kafkaFiles.CertificateAuthority).
						On(
							"CreateUser",
							username,
						).Return(kafkaFiles.Certificate, kafkaFiles.AccessKey, nil)
				})

				It("should not fail", func() {
					Expect(err).To(BeNil())
				})

				It("should create a user with the correct username", func() {
					mockKafka.AssertCalled(GinkgoT(), "CreateUser", username)
				})

				It("should create the 4 topics with correct names and retention", func() {
					mockKafka.AssertCalled(GinkgoT(), "CreateTopic", commandTopic, int64(-1))
					mockKafka.AssertCalled(GinkgoT(), "CreateTopic", changeTopic, int64(-1))
					mockKafka.AssertCalled(GinkgoT(), "CreateTopic", inputTopic, int64(-1))
					mockKafka.AssertCalled(GinkgoT(), "CreateTopic", receiptsTopic, int64(604800000))
				})

				It("should add the ACL's between the topics and the username", func() {
					mockKafka.AssertCalled(GinkgoT(), "AddACL", changeTopic, username, permission)
					mockKafka.AssertCalled(GinkgoT(), "AddACL", inputTopic, username, permission)
					mockKafka.AssertCalled(GinkgoT(), "AddACL", commandTopic, username, permission)
					mockKafka.AssertCalled(GinkgoT(), "AddACL", receiptsTopic, username, permission)
				})

				It("should write the credentials and config to the kafka-files configmap", func() {
					mockRepo.AssertCalled(GinkgoT(), "UpsertKafkaFiles", application, environment, kafkaFiles)
				})
			})

			When("given uppercase arguments", func() {
				BeforeEach(func() {
					customer = strings.ToUpper(customer)
					application = strings.ToUpper(application)
					environment = strings.ToUpper(environment)
					mockKafka.
						On(
							"GetCertificateAuthority",
						).Return(kafkaFiles.CertificateAuthority).
						On(
							"CreateUser",
							username,
						).Return(kafkaFiles.Certificate, kafkaFiles.AccessKey, nil)
				})
				It("should create a user with the correct lowercase username", func() {
					mockKafka.AssertCalled(GinkgoT(), "CreateUser", username)
				})

				It("should create the 4 topics with correct lowercase names", func() {
					mockKafka.AssertCalled(GinkgoT(), "CreateTopic", commandTopic, int64(-1))
					mockKafka.AssertCalled(GinkgoT(), "CreateTopic", changeTopic, int64(-1))
					mockKafka.AssertCalled(GinkgoT(), "CreateTopic", inputTopic, int64(-1))
					mockKafka.AssertCalled(GinkgoT(), "CreateTopic", receiptsTopic, int64(604800000))
				})
			})
		})
	})
})
