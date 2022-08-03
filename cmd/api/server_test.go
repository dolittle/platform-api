package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/containerregistry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"

	// "github.com/test-go/testify/mock"

	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"

	mockCR "github.com/dolittle/platform-api/mocks/pkg/platform/containerregistry"
	mockK8s "github.com/dolittle/platform-api/mocks/pkg/platform/k8s"
	mockStorage "github.com/dolittle/platform-api/mocks/pkg/platform/storage"
)

type response struct {
	status int
	field  string
	value  interface{}
}

type request struct {
	path   string
	secret string
}

func expect(req request, res response) {
	c := http.Client{}
	httpReq, _ := http.NewRequest("GET", req.path, nil)
	if req.secret == "" {
		httpReq.Header.Del("x-shared-secret")
	} else {
		httpReq.Header.Set("x-shared-secret", req.secret)
	}
	httpReq.Header.Set("Tenant-ID", "123")
	httpReq.Header.Set("User-ID", "666")

	httpRes, _ := c.Do(httpReq)
	Expect(httpRes).ToNot(BeNil())

	if res.status != 0 {
		Expect(httpRes.StatusCode).To(Equal(res.status))
	} else {
		Expect(httpRes.StatusCode).To(Equal(http.StatusOK))
	}

	if res.value != nil {
		body, _ := ioutil.ReadAll(httpRes.Body)
		var jsonData map[string]interface{}
		json.Unmarshal(body, &jsonData)
		fmt.Println(jsonData)
		Expect(jsonData[res.field]).To(Equal(res.value))
	}
}

var _ = Describe("Platform API", func() {
	var (
		logger                *logrus.Logger
		k8sPlatformRepoMock   *mockK8s.K8sPlatformRepo
		gitRepo               *mockStorage.Repo
		containerRegistryRepo *mockCR.ContainerRegistryRepo
		server                *httptest.Server
		path                  string
		now                   time.Time
	)

	BeforeEach(func() {
		viper.Set("tools.server.gitRepo.branch", "foo")
		viper.Set("tools.server.gitRepo.directory", "/tmp/dolittle_operation")
		viper.Set("tools.server.gitRepo.url", "git@github.com:dolittle-platform/Operations")
		viper.Set("tools.server.gitRepo.sshKey", "does/not/exist")
		viper.Set("tools.server.kubernetes.externalClusterHost", "external-host")
		viper.Set("tools.server.secret", "johnc")

		logger, _ = logrusTest.NewNullLogger()
		k8sPlatformRepoMock = new(mockK8s.K8sPlatformRepo)
		k8sPlatformRepoMock.On("GetSecret", mock.Anything, "12321", "acr").Return(&corev1.Secret{}, nil)
		k8sPlatformRepoMock.On("CanModifyApplicationWithResponse", mock.Anything, "123", "12321", "666").Return(true)
		gitRepo = new(mockStorage.Repo)
		gitRepo.On("GetTerraformTenant", mock.Anything).Return(platform.TerraformCustomer{ContainerRegistryName: "test1"}, nil)

		containerRegistryRepo = new(mockCR.ContainerRegistryRepo)
		k8sClient := fake.NewSimpleClientset()
		k8sRepoV2 := k8s.NewRepo(k8sClient, logger)
		k8sConfig := &rest.Config{}
		srv := NewServer(
			logger, gitRepo, k8sClient, k8sPlatformRepoMock, k8sRepoV2, k8sConfig, containerRegistryRepo)
		server = httptest.NewServer(srv.Handler)
	})

	AfterEach(func() {
		server.Close()
	})

	When("fetch images from Container Registry", func() {
		BeforeEach(func() {
			containerRegistryRepo.On("GetImages", mock.Anything).Return([]string{"helloworld"}, nil)
			path = fmt.Sprintf("%s/application/12321/containerregistry/images", server.URL)
		})

		It("should return 403 Forbidden when x-shared-secret header is not set", func() {
			expect(
				request{path: path, secret: ""},
				response{status: http.StatusForbidden},
			)
		})

		It("should return 403 Forbidden when x-shared-secret header is invalid", func() {
			expect(
				request{path: path, secret: "invalid header"},
				response{status: http.StatusForbidden},
			)
		})

		It("should include message in the response when x-shared-secret header is invalid", func() {
			expect(
				request{path: path, secret: "invalid header"},
				response{
					status: http.StatusForbidden,
					field:  "message",
					value:  "Shared secret is wrong",
				},
			)
		})

		It("should return container registry URL from the customer's config in our db (git storage)", func() {
			expect(
				request{path: path, secret: "johnc"},
				response{
					field: "url",
					value: "test1.azurecr.io",
				},
			)
		})

		It("should return the images", func() {
			expect(
				request{path: path, secret: "johnc"},
				response{
					field: "images",
					value: []interface{}{"helloworld"},
				},
			)
		})

	})

	When("fetch tags for the images", func() {
		BeforeEach(func() {
			containerRegistryRepo.On("GetTags", mock.Anything, "helloworld").Return(([]string{"latest", "v1"}), nil)

			path = fmt.Sprintf("%s/application/12321/containerregistry/tags/helloworld", server.URL)
		})

		It("should return the tags", func() {
			expect(
				request{path: path, secret: "johnc"},
				response{
					field: "tags",
					value: []interface{}{"latest", "v1"},
				},
			)
		})
	})

	When("fetch tags v2 for the images", func() {
		BeforeEach(func() {
			now = time.Now()
			containerRegistryRepo.On("GetImageTags", mock.Anything, "helloworld").Return([]containerregistry.ImageTag{{
				Name:           "label1",
				Digest:         "sha256:...",
				CreatedTime:    now,
				LastUpdateTime: now,
				Signed:         false,
			}}, nil)

			path = fmt.Sprintf("%s/application/12321/containerregistry/image-tags/helloworld", server.URL)
		})

		It("should return the tags with last modified date", func() {
			expect(
				request{path: path, secret: "johnc"},
				response{
					field: "tags",
					value: []interface{}{
						map[string]interface{}{
							"name":           "label1",
							"digest":         "sha256:...",
							"createdTime":    now.Format(time.RFC3339Nano),
							"lastUpdateTime": now.Format(time.RFC3339Nano),
							"signed":         false,
						},
					},
				},
			)
		})

		It("should return 403 Forbidden when x-shared-secret header is invalid", func() {
			expect(
				request{path: path, secret: "invalid header"},
				response{status: http.StatusForbidden},
			)
		})
	})
})
