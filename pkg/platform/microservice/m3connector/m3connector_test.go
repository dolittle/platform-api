package m3connector_test

import (
	"errors"
	"fmt"

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
		customer       string
		application    string
		environment    string
		resourcePrefix string
	)
	BeforeEach(func() {
		logger, _ := logrusTest.NewNullLogger()
		mockKafka = new(mockM3Connector.KafkaProvider)
		connector = m3connector.NewM3Connector(mockKafka, logger)
		customer = "test-customer"
		application = "test-application"
		environment = "test"
		resourcePrefix = fmt.Sprintf("cust_%s_%s_%s.m3", customer, application, environment)
	})
	Describe("Creating a new environmet", func() {
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
				resourcePrefix,
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
				resourcePrefix,
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
					resourcePrefix,
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
					resourcePrefix,
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
					resourcePrefix,
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
					resourcePrefix,
					permission,
				).Return(nil).
				On(
					"AddACL",
					inputTopic,
					resourcePrefix,
					permission,
				).Return(nil).
				On(
					"AddACL",
					commandTopic,
					resourcePrefix,
					permission,
				).Return(nil).
				On(
					"AddACL",
					receiptsTopic,
					resourcePrefix,
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
					resourcePrefix,
				).Return(nil).
				On(
					"CreateTopic",
					changeTopic,
					int64(-1),
				).Return(nil).
				On(
					"AddACL",
					changeTopic,
					resourcePrefix,
					permission,
				).Return(errors.New("error adding acl"))

			err := connector.CreateEnvironment(customer, application, environment)
			Expect(err).ToNot(BeNil())
			mock.AssertExpectationsForObjects(GinkgoT(), mockKafka)
		})
	})
})
