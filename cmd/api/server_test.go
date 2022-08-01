package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/dolittle/platform-api/cmd/api/mocks"
	"github.com/dolittle/platform-api/pkg/k8s"
	"github.com/dolittle/platform-api/pkg/platform/containerregistry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
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
	var gitRepo mocks.GitStorageRepoMock
	var containerRegistryRepo *mocks.ContainerRegistryMock
	var server *httptest.Server
	var paths map[string]string
	now := time.Now()

	BeforeEach(func() {
		viper.Set("tools.server.gitRepo.branch", "foo")
		viper.Set("tools.server.gitRepo.directory", "/tmp/dolittle_operation")
		viper.Set("tools.server.gitRepo.url", "git@github.com:dolittle-platform/Operations")
		viper.Set("tools.server.gitRepo.sshKey", "does/not/exist")
		viper.Set("tools.server.kubernetes.externalClusterHost", "external-host")
		viper.Set("tools.server.secret", "johnc")

		gitRepo = mocks.GitStorageRepoMock{}
		containerRegistryRepo = &mocks.ContainerRegistryMock{}
		containerRegistryRepo.StubAndReturnImages([]string{"helloworld"})
		containerRegistryRepo.StubAndReturnTags(([]string{"latest", "v1"}))
		containerRegistryRepo.StubAndReturnImageTags([]containerregistry.ImageTag{{
			Name:           "label1",
			Digest:         "sha256:...",
			CreatedTime:    now,
			LastUpdateTime: now,
			Signed:         false,
		}})

		logContext := logrus.StandardLogger()
		//k8sClient, k8sConfig := platformK8s.InitKubernetesClient()
		k8sClient := fake.NewSimpleClientset()
		k8sConfig := &rest.Config{}

		k8sPlatformRepoMock := &mocks.K8sPlatformRepoMock{}
		k8sPlatformRepoMock.StubGetSecretAndReturn(&corev1.Secret{})
		k8sRepoV2 := k8s.NewRepo(k8sClient, logContext.WithField("context", "k8s-repo-v2"))

		srv := NewServer(logContext, gitRepo, k8sClient, k8sPlatformRepoMock, k8sRepoV2, k8sConfig, containerRegistryRepo)
		server = httptest.NewServer(srv.Handler)

		paths = map[string]string{
			"images":     fmt.Sprintf("%s/application/12321/containerregistry/images", server.URL),
			"tags":       fmt.Sprintf("%s/application/12321/containerregistry/tags/helloworld", server.URL),
			"image-tags": fmt.Sprintf("%s/application/12321/containerregistry/image-tags/helloworld", server.URL),
		}
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("When fetch images from Container Registry", func() {

		It("should return 403 Forbidden when x-shared-secret header is not set", func() {
			expect(
				request{path: paths["images"], secret: ""},
				response{status: http.StatusForbidden},
			)
		})

		It("should rertun 403 Forbidden when x-shared-secret header is invalid", func() {
			expect(
				request{path: paths["images"], secret: "invalid header"},
				response{status: http.StatusForbidden},
			)
		})

		It("should include message in the response when x-shared-secret header is invalid", func() {
			expect(
				request{path: paths["images"], secret: "invalid header"},
				response{
					status: http.StatusForbidden,
					field:  "message",
					value:  "Shared secret is wrong",
				},
			)
		})

		It("should return container registry URL from the customer's config in our db (git storage)", func() {
			expect(
				request{path: paths["images"], secret: "johnc"},
				response{
					field: "url",
					value: "test1.azurecr.io",
				},
			)
		})

		It("should return the images", func() {
			expect(
				request{path: paths["images"], secret: "johnc"},
				response{
					field: "images",
					value: []interface{}{"helloworld"},
				},
			)
		})

	})

	Describe("When fetch tags for the images", func() {
		It("should return the tags", func() {
			expect(
				request{path: paths["tags"], secret: "johnc"},
				response{
					field: "tags",
					value: []interface{}{"latest", "v1"},
				},
			)
		})
	})

	Describe("When fetch tags v2 for the images", func() {
		It("should return the tags with last modified date", func() {
			expect(
				request{path: paths["image-tags"], secret: "johnc"},
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

		It("should rertun 403 Forbidden when x-shared-secret header is invalid", func() {
			expect(
				request{path: paths["image-tags"], secret: "invalid header"},
				response{status: http.StatusForbidden},
			)
		})
	})
})
