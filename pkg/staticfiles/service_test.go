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
		responseRecorder *httptest.ResponseRecorder
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

		responseRecorder = httptest.NewRecorder()
	})

	When("Getting file", func() {
		It("Success", func() {
			req := httptest.NewRequest(http.MethodGet, "/files/test.pdf", nil)
			storageProxy.On("Download", responseRecorder, "test.pdf")
			service.Get(responseRecorder, req)
			Expect(responseRecorder.Result().StatusCode).To(Equal(http.StatusOK))
		})
	})

	When("Remove a file", func() {
		It("Success", func() {
			wr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/manage/remove/files/test.pdf", nil)
			storageProxy.On("Delete", wr, req, "test.pdf")
			service.Remove(wr, req)
			Expect(responseRecorder.Result().StatusCode).To(Equal(http.StatusOK))
		})
	})

	When("Upload a file", func() {
		It("Success", func() {
			wr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/manage/add/test.pdf", nil)
			storageProxy.On("Upload", wr, req, "test.pdf")
			service.Upload(wr, req)
			Expect(responseRecorder.Result().StatusCode).To(Equal(http.StatusOK))
		})
	})
})
