package aiven_test

import (
	"github.com/dolittle/platform-api/pkg/aiven"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var client *aiven.Client

	BeforeEach(func() {
		client = aiven.NewClient("test-token", "test-project", "test-service")
	})

	Describe("Creating topics", func() {
		When("creating an empty topic", func() {
			It("should fail", func() {
				err := client.CreateTopic("", 1)

				Expect(err).ToNot(BeNil())
			})
		})
	})

	Describe("Creating users", func() {
		When("creating an user with an empty username", func() {
			It("should fail", func() {
				err := client.CreateUser("")

				Expect(err).ToNot(BeNil())
			})
		})
	})

	Describe("Creating ACL's", func() {
		When("creating an ACL with an empty topic", func() {
			It("should fail", func() {
				err := client.CreateACL("", "test-username", aiven.Read)

				Expect(err).ToNot(BeNil())
			})
		})

		When("creating an ACL with an empty username", func() {
			It("should fail", func() {
				err := client.CreateACL("test-topic", "", aiven.Read)

				Expect(err).ToNot(BeNil())
			})
		})
	})

	Describe("Creating an environment", func() {
		When("creating an environment", func() {
			It("should fail", func() {
				err := client.CreateACL("", "test-username", aiven.Read)

				Expect(err).ToNot(BeNil())
			})
		})

		When("creating an ACL with an empty username", func() {
			It("should fail", func() {
				err := client.CreateACL("test-topic", "", aiven.Read)

				Expect(err).ToNot(BeNil())
			})
		})
	})
})
