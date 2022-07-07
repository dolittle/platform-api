package application_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	mockK8s "github.com/dolittle/platform-api/mocks/pkg/k8s"
	mockApplication "github.com/dolittle/platform-api/mocks/pkg/platform/application"
	mockStorage "github.com/dolittle/platform-api/mocks/pkg/platform/storage"
	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform/application"
	jobK8s "github.com/dolittle/platform-api/pkg/platform/job/k8s"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	k8sSimple "github.com/dolittle/platform-api/pkg/platform/microservice/simple/k8s"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/dolittle/platform-api/pkg/platform/user"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

var _ = Describe("Testing Admin Application user endpoints", func() {
	var (
		request  *http.Request
		recorder *httptest.ResponseRecorder
		router   *mux.Router

		logger          *logrus.Logger
		gitRepo         *mockStorage.Repo
		clientSet       *fake.Clientset
		config          *rest.Config
		k8sRepo         platformK8s.K8sPlatformRepo
		userAccessRepo  *mockApplication.UserAccess
		roleBindingRepo *mockK8s.RepoRoleBinding
		service         application.Service

		customerID    string
		userID        string
		applicationID string
	)

	BeforeEach(func() {
		customerID = "fake-customer-123"
		userID = "fake-user-123"
		applicationID = "fake-application-123"
		subscriptionID := "TODO"
		externalClusterHost := "TODO"
		isProduction := false

		logger, _ = logrusTest.NewNullLogger()
		clientSet = fake.NewSimpleClientset()
		config = &rest.Config{}
		logger, _ = logrusTest.NewNullLogger()
		k8sRepo = platformK8s.NewK8sRepo(clientSet, config, logger)
		k8sRepoV2 := k8s.NewRepo(clientSet, logger.WithField("context", "k8s-repo-v2"))
		microserviceSimpleRepo := k8sSimple.NewSimpleRepo(clientSet, k8sRepo, k8sRepoV2, isProduction)
		userAccessRepo = &mockApplication.UserAccess{}
		roleBindingRepo = &mockK8s.RepoRoleBinding{}
		gitRepo = &mockStorage.Repo{}

		service = application.NewService(
			subscriptionID,
			externalClusterHost,
			clientSet,
			gitRepo,
			k8sRepo,
			jobK8s.CreateResourceConfig{},
			microserviceSimpleRepo,
			userAccessRepo,
			roleBindingRepo,
			logger.WithField("context", "application-service"),
		)

		recorder = httptest.NewRecorder()
		router = mux.NewRouter()
	})

	When("Listing user access", func() {
		BeforeEach(func() {
			testURL := fmt.Sprintf("/admin/customer/%s/application/%s/access/users", customerID, applicationID)
			router.HandleFunc("/admin/customer/{customerID}/application/{applicationID}/access/users", service.UserList)

			request = httptest.NewRequest(http.MethodGet, testURL, nil)
			request.Header.Add("User-ID", userID)
		})

		It("Failed to look up access due to error", func() {
			roleBindingRepo.On(
				"HasUserAdminAccess",
				userID,
			).Return(false, errors.New("fail"))

			router.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
		})

		It("Failed to look up access due to error", func() {
			roleBindingRepo.On(
				"HasUserAdminAccess",
				userID,
			).Return(false, nil)

			router.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusForbidden))
		})

		It("Failed to lookup application in storage", func() {
			roleBindingRepo.On(
				"HasUserAdminAccess",
				userID,
			).Return(true, nil)

			gitRepo.On(
				"GetApplication",
				customerID,
				applicationID,
			).Return(storage.JSONApplication{}, errors.New("fail"))

			router.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
		})

		It("Application not found in storage", func() {
			roleBindingRepo.On(
				"HasUserAdminAccess",
				userID,
			).Return(true, nil)

			gitRepo.On(
				"GetApplication",
				customerID,
				applicationID,
			).Return(storage.JSONApplication{}, storage.ErrNotFound)

			router.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusNotFound))
		})

		It("Failed to lookup users", func() {
			roleBindingRepo.On(
				"HasUserAdminAccess",
				userID,
			).Return(true, nil)

			gitRepo.On(
				"GetApplication",
				customerID,
				applicationID,
			).Return(storage.JSONApplication{
				ID:           applicationID,
				Name:         "fake-application",
				Environments: []storage.JSONEnvironment{},
			}, nil)

			userAccessRepo.On(
				"GetUsers",
				applicationID,
			).Return([]string{}, errors.New("fail"))

			router.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
		})

		Context("Successfully looking up users", func() {
			BeforeEach(func() {
				roleBindingRepo.On(
					"HasUserAdminAccess",
					userID,
				).Return(true, nil)

				gitRepo.On(
					"GetApplication",
					customerID,
					applicationID,
				).Return(storage.JSONApplication{
					ID:           applicationID,
					Name:         "fake-application",
					Environments: []storage.JSONEnvironment{},
				}, nil)
			})

			It("No users found", func() {
				userAccessRepo.On(
					"GetUsers",
					applicationID,
				).Return([]string{}, nil)

				router.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusOK))

				body, _ := io.ReadAll(recorder.Body)

				var expected application.HttpResponseAccessUsers
				json.Unmarshal(body, &expected)

				Expect(len(expected.Users)).To(Equal(0))
				Expect(expected.ID).To(Equal(applicationID))
				Expect(expected.Name).To(Equal("fake-application"))
			})

			It("3 users found", func() {
				userAccessRepo.On(
					"GetUsers",
					applicationID,
				).Return([]string{"chris", "joel", "pav"}, nil)

				router.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusOK))

				body, _ := io.ReadAll(recorder.Body)

				var expected application.HttpResponseAccessUsers
				json.Unmarshal(body, &expected)

				Expect(len(expected.Users)).To(Equal(3))
				Expect(expected.ID).To(Equal(applicationID))
				Expect(expected.Name).To(Equal("fake-application"))
			})
		})

	})

	When("Adding a user to an application", func() {
		BeforeEach(func() {
			testURL := fmt.Sprintf("/admin/customer/%s/application/%s/access/user", customerID, applicationID)
			router.HandleFunc("/admin/customer/{customerID}/application/{applicationID}/access/user", service.UserAdd)
			request = httptest.NewRequest(http.MethodPost, testURL, nil)
			request.Header.Add("User-ID", userID)
		})

		It("Failed to look up access due to error", func() {
			roleBindingRepo.On(
				"HasUserAdminAccess",
				userID,
			).Return(false, errors.New("fail"))

			router.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
		})

		It("Failed to look up access due to error", func() {
			roleBindingRepo.On(
				"HasUserAdminAccess",
				userID,
			).Return(false, nil)

			router.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusForbidden))
		})

		It("Bad payload", func() {
			r := io.NopCloser(strings.NewReader("hello world"))
			request.Body = r

			roleBindingRepo.On(
				"HasUserAdminAccess",
				userID,
			).Return(true, nil)

			router.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})

		Context("Adding user", func() {
			var (
				email string
			)
			BeforeEach(func() {
				email = "human@dolittle.com"
				input := application.HttpInputAccessUser{
					Email: email,
				}

				jsonPayload, _ := json.Marshal(input)
				r := io.NopCloser(bytes.NewBuffer(jsonPayload))
				request.Body = r

				roleBindingRepo.On(
					"HasUserAdminAccess",
					userID,
				).Return(true, nil)
			})

			It("Email not found", func() {
				userAccessRepo.On(
					"AddUser",
					customerID,
					applicationID,
					email,
				).Return(user.ErrNotFound)

				router.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
			})

			It("Too many results (azure specific)", func() {
				userAccessRepo.On(
					"AddUser",
					customerID,
					applicationID,
					email,
				).Return(user.ErrTooManyResults)

				router.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusUnprocessableEntity))
			})

			It("Email already in the application group", func() {
				userAccessRepo.On(
					"AddUser",
					customerID,
					applicationID,
					email,
				).Return(user.ErrEmailAlreadyExists)

				router.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusUnprocessableEntity))
			})

			It("Success", func() {
				userAccessRepo.On(
					"AddUser",
					customerID,
					applicationID,
					email,
				).Return(nil)

				router.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusOK))
			})

		})
	})

	When("Removing a user from an application", func() {
		BeforeEach(func() {
			testURL := fmt.Sprintf("/admin/customer/%s/application/%s/access/user", customerID, applicationID)
			router.HandleFunc("/admin/customer/{customerID}/application/{applicationID}/access/user", service.UserRemove)
			request = httptest.NewRequest(http.MethodDelete, testURL, nil)
			request.Header.Add("User-ID", userID)
		})

		It("Failed to look up access due to error", func() {
			roleBindingRepo.On(
				"HasUserAdminAccess",
				userID,
			).Return(false, errors.New("fail"))

			router.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
		})

		It("Failed to look up access due to error", func() {
			roleBindingRepo.On(
				"HasUserAdminAccess",
				userID,
			).Return(false, nil)

			router.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusForbidden))
		})

		It("Bad payload", func() {
			r := io.NopCloser(strings.NewReader("hello world"))
			request.Body = r

			roleBindingRepo.On(
				"HasUserAdminAccess",
				userID,
			).Return(true, nil)

			router.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})

		Context("Removing user", func() {
			var (
				email string
			)
			BeforeEach(func() {
				email = "human@dolittle.com"
				input := application.HttpInputAccessUser{
					Email: email,
				}

				jsonPayload, _ := json.Marshal(input)
				r := io.NopCloser(bytes.NewBuffer(jsonPayload))
				request.Body = r

				roleBindingRepo.On(
					"HasUserAdminAccess",
					userID,
				).Return(true, nil)
			})

			It("Email not found", func() {
				userAccessRepo.On(
					"RemoveUser",
					applicationID,
					email,
				).Return(user.ErrNotFound)

				router.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
			})

			It("Too many results (azure specific)", func() {
				userAccessRepo.On(
					"RemoveUser",
					applicationID,
					email,
				).Return(user.ErrTooManyResults)

				router.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusUnprocessableEntity))
			})

			It("Success", func() {
				userAccessRepo.On(
					"RemoveUser",
					applicationID,
					email,
				).Return(nil)

				router.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(http.StatusOK))
			})
		})
	})
})
