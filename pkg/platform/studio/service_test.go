package studio

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	mockStorage "github.com/dolittle/platform-api/mocks/pkg/platform/storage"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("Studio service", func() {
	var (
		logger       *logrus.Logger
		mockRepo     *mockStorage.Repo
		service      service
		recorder     *httptest.ResponseRecorder
		router       *mux.Router
		studioConfig platform.StudioConfig
		customerID   string
	)

	BeforeEach(func() {
		logger, _ = logrusTest.NewNullLogger()
		mockRepo = new(mockStorage.Repo)
		service = NewService(mockRepo, logger)
		recorder = httptest.NewRecorder()
		router = mux.NewRouter()
		router.HandleFunc("/studio/customer/{customerID}", service.Get)
		studioConfig = platform.StudioConfig{
			BuildOverwrite:       false,
			DisabledEnvironments: []string{"*"},
			CanCreateApplication: false,
		}
		customerID = "4fd6927e-f5cf-44f8-9252-4058f5f24d6d"

	})

	When("getting a single customers studio configuration", func() {
		It("returns a single studio configuration", func() {
			request, _ := http.NewRequest("GET", fmt.Sprintf("/studio/customer/%s", customerID), nil)

			mockRepo.On(
				"GetStudioConfig",
				customerID,
			).Return(
				func(customerID string) platform.StudioConfig {
					return studioConfig
				},
				func(customerID string) error {
					return nil
				},
			)

			router.ServeHTTP(recorder, request)

			Expect(recorder.Code).To(Equal(http.StatusOK))

			bytes, _ := json.Marshal(studioConfig)
			Expect(recorder.Body.String()).To(Equal(string(bytes)))
		})

		It("returns a 500 if it fails to get the configuration", func() {
			request, _ := http.NewRequest("GET", fmt.Sprintf("/studio/customer/%s", customerID), nil)

			mockRepo.On(
				"GetStudioConfig",
				customerID,
			).Return(
				func(customerID string) platform.StudioConfig {
					return platform.StudioConfig{}
				},
				func(customerID string) error {
					return errors.New("oh nyo something went ^w^ tewwibwy wwong")
				},
			)

			router.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
		})
	})
})
