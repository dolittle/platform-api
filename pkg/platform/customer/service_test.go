package customer

import (
	"github.com/dolittle/platform-api/pkg/platform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("Customer service", func() {

	When("getting a single customer", func() {
		var (
			logger              *logrus.Logger
			mockRepo            *mockStorage.Repo
			mockRoleBindingRepo *mockK8s.RepoRoleBinding
			service             service
			studioConfig        platform.StudioConfig
			customerConfig      platform.Customer
			customerID          string
			jsonConfig          HTTPStudioConfig
			userID              string
		)
		BeforeEach(func() {
			logger, _ = logrusTest.NewNullLogger()
			mockRepo = new(mockStorage.Repo)
			mockRoleBindingRepo = new(mockK8s.RepoRoleBinding)
			service = NewService(mockRepo, logger, mockRoleBindingRepo)
			customerID = "4fd6927e-f5cf-44f8-9252-4058f5f24d6d"
			userID = "ad352a4f-d4a1-45a8-9db8-c1ce1a018981"
		})

		Context("and the user has access", func() {
			It("returns a single customer successfully", func() {

			})

			It("fails if the customer.json doesn't exist", func() {

			})

			It("fails if the studio.json doesn't exist", func() {

			})
		})

		Context("and the user doesn't have access", func() {
			It("returns a rejection", func() {

			})
		})

		It("should fail if the access check fails", func() {

		})
	})
})
