package application_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	mockApplication "github.com/dolittle/platform-api/mocks/pkg/platform/application"
	mockStorage "github.com/dolittle/platform-api/mocks/pkg/platform/storage"
	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/application"
	jobK8s "github.com/dolittle/platform-api/pkg/platform/job/k8s"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	k8sSimple "github.com/dolittle/platform-api/pkg/platform/microservice/simple/k8s"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

var _ = Describe("Testing endpoints", func() {
	var (
		req *http.Request
		w   *httptest.ResponseRecorder

		logger    *logrus.Logger
		gitRepo   *mockStorage.Repo
		clientSet *fake.Clientset
		config    *rest.Config
		k8sRepo   platformK8s.K8sRepo
		service   application.Service

		customerID    string
		applicationID string
	)

	BeforeEach(func() {
		customerID = "fake-customer-123"
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
		userAccessRepo := &mockApplication.UserAccess{}

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
			k8sRepoV2,
			logger.WithField("context", "application-service"),
		)
	})

	When("Creating an application", func() {
		It("check name", func() {
			name := "hello world"
			valid := application.IsApplicationNameValid(name)
			Expect(valid).To(BeFalse())
			name = "helloworld1"
			valid = application.IsApplicationNameValid(name)
			Expect(valid).To(BeTrue())
		})
		It("Studio config is missing", func() {
			want := errors.New("fail")
			gitRepo.On(
				"GetStudioConfig",
				customerID,
			).Return(platform.StudioConfig{}, want)

			url := "http://studio/application"
			req = httptest.NewRequest("POST", url, nil)
			w = httptest.NewRecorder()

			req.Header.Set("Tenant-ID", customerID)

			service.Create(w, req)
			resp := w.Result()
			Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))

			body, _ := io.ReadAll(resp.Body)

			var response utils.HTTPMessageResponse
			json.Unmarshal(body, &response)
			Expect(response.Message).To(Equal(want.Error()))
		})

		It("Studio has creation of applications disabled", func() {
			gitRepo.On(
				"GetStudioConfig",
				customerID,
			).Return(platform.StudioConfig{
				CanCreateApplication: false,
			}, nil)

			url := "http://studio/application"
			req = httptest.NewRequest("POST", url, nil)
			w = httptest.NewRecorder()

			req.Header.Set("Tenant-ID", customerID)

			service.Create(w, req)
			resp := w.Result()
			Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
		})
	})
	When("GetApplications", func() {
		It("Has 1 application with 2 environments", func() {
			gitRepo.On(
				"GetStudioConfig",
				customerID,
			).Return(platform.StudioConfig{
				CanCreateApplication: true,
			}, nil)

			gitRepo.On(
				"GetTerraformTenant",
				customerID,
			).Return(platform.TerraformCustomer{
				GUID: customerID,
				Name: "fake-customer",
			}, nil)

			applicationName := "fake-application"
			gitRepo.On(
				"GetApplications",
				customerID,
			).Return([]storage.JSONApplication{
				{
					ID:   applicationID,
					Name: applicationName,
					Environments: []storage.JSONEnvironment{
						{
							Name: "Dev",
						},
						{
							Name: "Prod",
						},
					},
				},
			}, nil)

			url := "http://studio/applications"
			req = httptest.NewRequest("GET", url, nil)
			w = httptest.NewRecorder()

			req.Header.Set("Tenant-ID", customerID)

			service.GetApplications(w, req)
			resp := w.Result()
			body, _ := io.ReadAll(resp.Body)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			var response application.HttpResponseApplications
			json.Unmarshal(body, &response)

			Expect(response.ID).To(Equal(customerID))
			Expect(response.CanCreateApplication).To(Equal(true))

			Expect(len(response.Applications)).To(Equal(2))
			Expect(response.Applications[0].ID).To(Equal(applicationID))
			Expect(response.Applications[0].Name).To(Equal(applicationName))
			Expect(response.Applications[0].Environment).To(Equal("Dev"))
			Expect(response.Applications[1].ID).To(Equal(applicationID))
			Expect(response.Applications[1].Name).To(Equal(applicationName))
			Expect(response.Applications[1].Environment).To(Equal("Prod"))
		})

	})
	When("GetByID", func() {
		It("Missing studio config", func() {
			want := errors.New("fail")
			gitRepo.On(
				"GetStudioConfig",
				customerID,
			).Return(platform.StudioConfig{}, want)

			url := fmt.Sprintf(`http://studio/application/%s`, applicationID)
			req := httptest.NewRequest("GET", url, nil)

			w := httptest.NewRecorder()

			vars := map[string]string{
				"applicationID": applicationID,
			}

			req.Header.Set("Tenant-ID", customerID)
			req = mux.SetURLVars(req, vars)

			service.GetByID(w, req)
			resp := w.Result()
			Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
		})

		When("Successfully found", func() {
			BeforeEach(func() {
				gitRepo.On(
					"GetStudioConfig",
					customerID,
				).Return(platform.StudioConfig{}, nil)

				gitRepo.On(
					"GetTerraformTenant",
					customerID,
				).Return(platform.TerraformCustomer{
					GUID: customerID,
					Name: "fake-customer",
				}, nil)

				gitRepo.On(
					"GetTerraformApplication",
					customerID,
					applicationID,
				).Return(platform.TerraformApplication{
					ApplicationID: applicationID,
					Name:          "fake-application",
				}, nil)

				url := fmt.Sprintf(`http://studio/application/%s`, applicationID)
				req = httptest.NewRequest("GET", url, nil)
				w = httptest.NewRecorder()

				vars := map[string]string{
					"applicationID": applicationID,
				}

				req.Header.Set("Tenant-ID", customerID)
				req = mux.SetURLVars(req, vars)
			})

			It("Empty environments", func() {
				gitRepo.On(
					"GetApplication",
					customerID,
					applicationID,
				).Return(storage.JSONApplication{
					ID:           applicationID,
					Name:         "fake-application",
					Environments: []storage.JSONEnvironment{},
				}, nil)

				gitRepo.On(
					"GetMicroservices",
					customerID,
					applicationID,
				).Return([]platform.HttpMicroserviceBase{}, nil)

				service.GetByID(w, req)
				resp := w.Result()
				body, _ := io.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				var response application.HttpResponseApplication
				json.Unmarshal(body, &response)

				Expect(len(response.Environments)).To(Equal(0))
			})

			It("2 environments", func() {
				gitRepo.On(
					"GetApplication",
					customerID,
					applicationID,
				).Return(storage.JSONApplication{
					ID:   applicationID,
					Name: "fake-application",
					Environments: []storage.JSONEnvironment{
						{
							Name: "Dev",
						},
						{
							Name: "Prod",
						},
					},
				}, nil)

				gitRepo.On(
					"GetMicroservices",
					customerID,
					applicationID,
				).Return([]platform.HttpMicroserviceBase{}, nil)

				gitRepo.On(
					"IsAutomationEnabledWithStudioConfig",
					mock.Anything,
					applicationID,
					mock.Anything,
				).Return(true).Twice()

				service.GetByID(w, req)
				resp := w.Result()
				body, _ := io.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				fmt.Println(resp.Header.Get("Content-Type"))
				var response application.HttpResponseApplication
				json.Unmarshal(body, &response)

				Expect(len(response.Environments)).To(Equal(2))
				Expect(response.Environments[0].Name).To(Equal("Dev"))
				Expect(response.Environments[0].AutomationEnabled).To(BeTrue())
				Expect(response.Environments[1].Name).To(Equal("Prod"))
				Expect(response.Environments[1].AutomationEnabled).To(BeTrue())
			})
		})
	})

})
