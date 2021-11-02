package staticfiles_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/dolittle-entropy/platform-api/pkg/staticfiles"
)

var _ = Describe("Service", func() {
	var (
		service staticfiles.Service
		logger  *logrus.Logger
		//hook    *test.Hook
	)
	BeforeEach(func() {
		//logger, hook = test.NewNullLogger()
		logger, _ = test.NewNullLogger()
		uriPrefix := ""
		tenantID := "fake-123"
		var storageProxy staticfiles.StorageProxy
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
			req := httptest.NewRequest(http.MethodGet, "/sloth", nil)

			service.Get(wr, req)
			Expect(wr.Result().StatusCode).To(Equal(http.StatusOK))
		})
	})

})
