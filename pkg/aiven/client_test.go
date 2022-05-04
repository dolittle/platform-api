package aiven

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var client *client

	BeforeEach(func() {
		client = NewClient("test-token", "test-project", "test-service")
	})

	Describe("Creating topics", func() {
		When("creating an empty topic", func() {
			It("should fail", func() {
				_, err := client.CreateTopic("", 0)

				Expect(err).ToNot(BeNil())
			})
		})
	})

	Describe("Creating users", func() {
		When("creating an user with an empty username", func() {
			It("should fail", func() {
				_, err := client.CreateUser("")

				Expect(err).ToNot(BeNil())
			})
		})
	})

	Describe("Creating ACL's", func() {
		When("creating an ACL with an empty topic", func() {
			It("should fail", func() {
				_, err := client.CreateACL("", "test-username", Read)

				Expect(err).ToNot(BeNil())
			})
		})

		When("creating an ACL with an empty username", func() {
			It("should fail", func() {
				_, err := client.CreateACL("test-topic", "", Read)

				Expect(err).ToNot(BeNil())
			})
		})
	})

	Describe("Creating an environment", func() {
		When("creating an environment", func() {
			It("should fail", func() {
				_, err := client.CreateACL("", "test-username", Read)

				Expect(err).ToNot(BeNil())
			})
		})

		When("creating an ACL with an empty username", func() {
			It("should fail", func() {
				_, err := client.CreateACL("test-topic", "", Read)

				Expect(err).ToNot(BeNil())
			})
		})
	})
})
