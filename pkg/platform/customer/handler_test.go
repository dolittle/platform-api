package customer_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"

	"github.com/dolittle/platform-api/pkg/platform/customer"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("Handler", func() {
	var (
		logger      *logrus.Logger
		recorder    *httptest.ResponseRecorder
		mockService customer.CustomerProvider
		router      *mux.Router
		customerID  string
		customerUrl string
		userID      string
	)
	handler := customer.NewHandler(mockService, logger)

	BeforeEach(func() {
		logger, _ = logrusTest.NewNullLogger()
		recorder = httptest.NewRecorder()
		router = mux.NewRouter()
		customerID = "4fd6927e-f5cf-44f8-9252-4058f5f24d6d"
		customerUrl = fmt.Sprintf("/studio/customer/%s", customerID)
		userID = "ad352a4f-d4a1-45a8-9db8-c1ce1a018981"
	})

	When("creating a customer", func() {
		Context("and it works", func() {
			It("should return a 200", func() {
				customerInput := customer.HttpCustomerInput{
					Name: "test",
				}

				customerPayload, _ := json.Marshal(customerInput)

				request := httptest.NewRequest("POST", customerUrl, bytes.NewBuffer(customerPayload))
				router.HandleFunc("/studio/customer/{customerID}", handler.Create)
				request := httptest.NewRequest("GET", customerUrl, nil)
				request.Header.Add("User-ID", userID)

			})
		})
	})
})
