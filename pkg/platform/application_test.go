package platform_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/dolittle/platform-api/pkg/platform"
)

var _ = Describe("Application", func() {
	var (
		application platform.HttpResponseApplication
	)
	BeforeEach(func() {
		application = platform.HttpResponseApplication{
			ID:           "9b5c27c5-3e7c-4d3b-9e5f-df3401b6d61d",
			Name:         "Some application",
			TenantID:     "16a57ea2-0142-45f1-9c5b-e92705b9cbfb",
			TenantName:   "Some tenant",
			Environments: make([]platform.HttpInputEnvironment, 0),
		}
	})
	Describe("for GetTenantForEnvironment", func() {
		var (
			resultTenant platform.TenantId
			resultErr    error
		)
		Describe("and there are no environments", func() {
			BeforeEach(func() {
				_, resultErr = application.GetTenantForEnvironment("Prod")
			})
			It("should fail because there are no configured environments", func() {
				var expectedErr *platform.NoConfiguredEnvironmentsError
				Expect(errors.As(resultErr, &expectedErr)).To(BeTrue())
			})
		})
		Describe("and environment is not found", func() {
			BeforeEach(func() {
				application.Environments = append(application.Environments, platform.HttpInputEnvironment{
					Name: "Dev",
				})
				_, resultErr = application.GetTenantForEnvironment("Prod")
			})
			It("should fail because environment was not found", func() {
				var expectedErr *platform.EnvironmentNotFoundError
				Expect(errors.As(resultErr, &expectedErr)).To(BeTrue())
			})
		})
		Describe("and there are no configured tenants for environment", func() {
			BeforeEach(func() {
				application.Environments = append(application.Environments, platform.HttpInputEnvironment{
					Name:    "Prod",
					Tenants: make([]platform.TenantId, 0),
				})
				_, resultErr = application.GetTenantForEnvironment("Prod")
			})
			It("should fail because environment was not found", func() {
				var expectedErr *platform.NoConfiguredTenantsError
				Expect(errors.As(resultErr, &expectedErr)).To(BeTrue())
			})
		})
		Describe("and there is one configured tenant for environment", func() {
			var configuredTenant platform.TenantId
			BeforeEach(func() {
				configuredTenant = "8e63a95c-e24b-44d2-a19b-0bebdc8a0832"
				application.Environments = append(application.Environments, platform.HttpInputEnvironment{
					Name:    "Prod",
					Tenants: append(make([]platform.TenantId, 0), configuredTenant),
				})
				resultTenant, resultErr = application.GetTenantForEnvironment("Prod")
			})
			It("should not fail", func() {
				Expect(resultErr).To(BeNil())
			})
			It("should get the configured tenant", func() {
				Expect(resultTenant).To(Equal(configuredTenant))
			})
		})
		Describe("and there are multiple configured tenants for environment", func() {
			var (
				firstConfiguredTenant  platform.TenantId
				secondConfiguredTenant platform.TenantId
				thirdConfiguredTenant  platform.TenantId
			)
			BeforeEach(func() {
				firstConfiguredTenant = "8e63a95c-e24b-44d2-a19b-0bebdc8a0832"
				secondConfiguredTenant = "a7c5eb38-741b-4d90-b33a-5cad0ee9d52c"
				thirdConfiguredTenant = "ff2d4862-5405-476d-91f5-331c83872687"
				application.Environments = append(application.Environments, platform.HttpInputEnvironment{
					Name:    "Prod",
					Tenants: append(make([]platform.TenantId, 0), firstConfiguredTenant, secondConfiguredTenant, thirdConfiguredTenant),
				})
				resultTenant, resultErr = application.GetTenantForEnvironment("Prod")
			})
			It("should not fail", func() {
				Expect(resultErr).To(BeNil())
			})
			It("should get the first configured tenant", func() {
				Expect(resultTenant).To(Equal(firstConfiguredTenant))
			})
		})
	})
})
