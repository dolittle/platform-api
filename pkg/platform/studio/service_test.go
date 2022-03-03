package studio

import (
	"bytes"
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
	"github.com/stretchr/testify/mock"
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
		customerUrl  string
		jsonConfig   HTTPStudioConfig
	)

	BeforeEach(func() {
		logger, _ = logrusTest.NewNullLogger()
		mockRepo = new(mockStorage.Repo)
		service = NewService(mockRepo, logger)
		recorder = httptest.NewRecorder()
		router = mux.NewRouter()
		customerID = "4fd6927e-f5cf-44f8-9252-4058f5f24d6d"
		customerUrl = fmt.Sprintf("/studio/customer/%s", customerID)
	})

	When("getting a single customers studio configuration", func() {
		BeforeEach(func() {
			router.HandleFunc("/studio/customer/{customerID}", service.Get)
			studioConfig = platform.StudioConfig{
				BuildOverwrite:       false,
				DisabledEnvironments: []string{"*"},
				CanCreateApplication: false,
			}
			jsonConfig = HTTPStudioConfig{
				BuildOverwrite:       false,
				DisabledEnvironments: []string{"*"},
				CanCreateApplication: false,
			}
		})

		It("returns a single studio configuration", func() {
			request, _ := http.NewRequest("GET", customerUrl, nil)

			mockRepo.On(
				"GetStudioConfig",
				customerID,
			).Return(studioConfig, nil)

			router.ServeHTTP(recorder, request)

			Expect(recorder.Code).To(Equal(http.StatusOK))

			bytes, _ := json.Marshal(jsonConfig)
			Expect(recorder.Body.String()).To(Equal(string(bytes)))
			mock.AssertExpectationsForObjects(GinkgoT(), mockRepo)
		})

		It("returns a 500 if it fails to get the configuration", func() {
			request, _ := http.NewRequest("GET", customerUrl, nil)

			mockRepo.On(
				"GetStudioConfig",
				customerID,
			).Return(platform.StudioConfig{}, errors.New("oh nyo something went ^w^ tewwibwy wwong"))

			router.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			mock.AssertExpectationsForObjects(GinkgoT(), mockRepo)
		})
	})

	When("saving a studio config", func() {
		BeforeEach(func() {
			router.HandleFunc("/studio/customer/{customerID}", service.Save)
			studioConfig = platform.StudioConfig{
				BuildOverwrite:       false,
				DisabledEnvironments: []string{"*"},
				CanCreateApplication: false,
			}
			jsonConfig = HTTPStudioConfig{
				BuildOverwrite:       false,
				DisabledEnvironments: []string{"*"},
				CanCreateApplication: false,
			}
			customerID = "4fd6927e-f5cf-44f8-9252-4058f5f24d6d"
		})

		It("should save the given studio config", func() {
			jsonPayload, _ := json.Marshal(jsonConfig)

			request, _ := http.NewRequest("POST", customerUrl, bytes.NewBuffer(jsonPayload))

			mockRepo.On(
				"SaveStudioConfig",
				customerID,
				studioConfig,
			).Return(nil)

			router.ServeHTTP(recorder, request)

			Expect(recorder.Code).To(Equal(http.StatusOK))
			mock.AssertExpectationsForObjects(GinkgoT(), mockRepo)
		})

		It("should return a 400 if the given config payload is wrong", func() {
			wrongPayload := []byte(`{"imnot": "studioconfig"}`)

			request, _ := http.NewRequest("POST", customerUrl, bytes.NewBuffer(wrongPayload))

			mockRepo.On(
				"SaveStudioConfig",
				customerID,
				mock.Anything,
			).Return(nil)

			router.ServeHTTP(recorder, request)

			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			mockRepo.AssertNotCalled(GinkgoT(), "SaveStudioConfig", mock.Anything, mock.Anything)
		})

		It("should return a 500 if the config saving fails", func() {
			jsonPayload, _ := json.Marshal(jsonConfig)

			request, _ := http.NewRequest("POST", customerUrl, bytes.NewBuffer(jsonPayload))

			mockRepo.On(
				"SaveStudioConfig",
				customerID,
				studioConfig,
			).Return(errors.New("expect THIS!"))

			router.ServeHTTP(recorder, request)

			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			mock.AssertExpectationsForObjects(GinkgoT(), mockRepo)
		})
	})
})
