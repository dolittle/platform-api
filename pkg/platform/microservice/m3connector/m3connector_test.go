package m3connector_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/dolittle/platform-api/pkg/platform/microservice/m3connector"

	mockM3Connector "github.com/dolittle/platform-api/mocks/pkg/platform/microservice/m3connector"
)

var _ = Describe("M3connector", func() {
	var (
		connector   *m3connector.M3Connector
		mockKafka   *mockM3Connector.KafkaProvider
		customer    string
		application string
		environment string
	)
	BeforeEach(func() {
		mockKafka = new(mockM3Connector.KafkaProvider)
		connector = m3connector.NewM3Connector(mockKafka)
		customer = "test-customer"
		application = "test-application"
		environment = "test"
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
			username := fmt.Sprintf("cust_%s_%s_%s.m3", customer, application, environment)

			mockKafka.On(
				"CreateUser",
				username,
			).Return(nil)

			err := connector.CreateEnvironment(customer, application, environment)
			Expect(err).To(BeNil())
			mock.AssertExpectationsForObjects(GinkgoT(), mockKafka)
		})
	})
})
