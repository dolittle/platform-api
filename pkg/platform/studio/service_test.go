package studio

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	mockK8s "github.com/dolittle/platform-api/mocks/pkg/k8s"
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
		logger              *logrus.Logger
		mockRepo            *mockStorage.Repo
		mockRoleBindingRepo *mockK8s.RepoRoleBinding
		service             service
		recorder            *httptest.ResponseRecorder
		router              *mux.Router
		studioConfig        platform.StudioConfig
		customerID          string
		customerUrl         string
		jsonConfig          HTTPStudioConfig
		userID              string
	)

	BeforeEach(func() {
		logger, _ = logrusTest.NewNullLogger()
		mockRepo = new(mockStorage.Repo)
		mockRoleBindingRepo = new(mockK8s.RepoRoleBinding)
		service = NewService(mockRepo, logger, mockRoleBindingRepo)
		recorder = httptest.NewRecorder()
		router = mux.NewRouter()
		customerID = "4fd6927e-f5cf-44f8-9252-4058f5f24d6d"
		customerUrl = fmt.Sprintf("/studio/customer/%s", customerID)
		userID = "ad352a4f-d4a1-45a8-9db8-c1ce1a018981"
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
			request := httptest.NewRequest("GET", customerUrl, nil)
			request.Header.Add("User-ID", userID)

			mockRoleBindingRepo.On(
				"HasUserAdminAccess",
				userID,
			).Return(true, nil)

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
			request := httptest.NewRequest("GET", customerUrl, nil)
			request.Header.Add("User-ID", userID)

			mockRoleBindingRepo.On(
				"HasUserAdminAccess",
				userID,
			).Return(true, nil)

			mockRepo.On(
				"GetStudioConfig",
				customerID,
			).Return(platform.StudioConfig{}, errors.New("oh nyo something went ^w^ tewwibwy wwong"))

			router.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			mock.AssertExpectationsForObjects(GinkgoT(), mockRepo)
		})

		It("returns a 403 if the user doesn't have admin access", func() {
			request := httptest.NewRequest("GET", customerUrl, nil)
			request.Header.Add("User-ID", "nonexistant-user-id")

			mockRoleBindingRepo.On(
				"HasUserAdminAccess",
				mock.Anything,
			).Return(false, nil)

			router.ServeHTTP(recorder, request)

			Expect(recorder.Code).To(Equal(http.StatusForbidden))
		})

		It("returns a 500 if the admin check fails", func() {
			request := httptest.NewRequest("GET", customerUrl, nil)
			request.Header.Add("User-ID", "nonexistant-user-id")

			mockRoleBindingRepo.On(
				"HasUserAdminAccess",
				mock.Anything,
			).Return(false, errors.New("expected this"))

			router.ServeHTTP(recorder, request)

			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
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

			request := httptest.NewRequest("POST", customerUrl, bytes.NewBuffer(jsonPayload))
			request.Header.Add("User-ID", userID)

			mockRoleBindingRepo.On(
				"HasUserAdminAccess",
				userID,
			).Return(true, nil)

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

			request := httptest.NewRequest("POST", customerUrl, bytes.NewBuffer(wrongPayload))
			request.Header.Add("User-ID", userID)

			mockRoleBindingRepo.On(
				"HasUserAdminAccess",
				userID,
			).Return(true, nil)

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

			request := httptest.NewRequest("POST", customerUrl, bytes.NewBuffer(jsonPayload))
			request.Header.Add("User-ID", userID)

			mockRoleBindingRepo.On(
				"HasUserAdminAccess",
				userID,
			).Return(true, nil)

			mockRepo.On(
				"SaveStudioConfig",
				customerID,
				studioConfig,
			).Return(errors.New("expect THIS!"))

			router.ServeHTTP(recorder, request)

			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			mock.AssertExpectationsForObjects(GinkgoT(), mockRepo)
		})

		It("returns a 403 if the user doesn't have admin access", func() {
			jsonPayload, _ := json.Marshal(jsonConfig)

			request := httptest.NewRequest("POST", customerUrl, bytes.NewBuffer(jsonPayload))
			request.Header.Add("User-ID", "nonexistant-user-id")

			mockRoleBindingRepo.On(
				"HasUserAdminAccess",
				mock.Anything,
			).Return(false, nil)

			mockRepo.On(
				"SaveStudioConfig",
				customerID,
				mock.Anything,
			).Return(nil)

			router.ServeHTTP(recorder, request)

			Expect(recorder.Code).To(Equal(http.StatusForbidden))
			// check that the save wasn't called if unauthorized
			mockRepo.AssertNotCalled(GinkgoT(), "SaveStudioConfig", mock.Anything, mock.Anything)
		})

		It("returns a 500 if the admin check fails", func() {
			request := httptest.NewRequest("POST", customerUrl, nil)
			request.Header.Add("User-ID", "nonexistant-user-id")

			mockRoleBindingRepo.On(
				"HasUserAdminAccess",
				mock.Anything,
			).Return(false, errors.New("expected this"))

			router.ServeHTTP(recorder, request)

			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
		})
	})
})
