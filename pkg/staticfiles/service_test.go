package staticfiles_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/dolittle-entropy/platform-api/pkg/staticfiles"
	"github.com/dolittle-entropy/platform-api/pkg/staticfiles/mocks"
)

var _ = Describe("Service", func() {
	var (
		storageProxy *mocks.StorageProxy
		service      staticfiles.Service
		logger       *logrus.Logger
		//hook    *test.Hook
	)
	BeforeEach(func() {
		//logger, hook = test.NewNullLogger()
		logger, _ = test.NewNullLogger()
		uriPrefix := "/files/"
		tenantID := "fake-123"

		storageProxy = &mocks.StorageProxy{}
		service = staticfiles.NewService(
			logger,
			uriPrefix,
			storageProxy,
			tenantID,
		)
	})

	When("Getting file", func() {
		It("Success", func() {
			wr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/files/test.pdf", nil)
			storageProxy.On("Download", wr, "test.pdf")
			service.Get(wr, req)
			Expect(wr.Result().StatusCode).To(Equal(http.StatusOK))
		})
	})

})
