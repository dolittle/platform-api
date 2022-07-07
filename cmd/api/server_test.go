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

func createRequest(path, xSharedSecret string) *http.Request {
	request, _ := http.NewRequest("GET", path, nil)
	if xSharedSecret == "" {
		request.Header.Del("x-shared-secret")
	} else {
		request.Header.Set("x-shared-secret", xSharedSecret)
	}
	request.Header.Set("Tenant-ID", "123")
	request.Header.Set("User-ID", "666")
	return request
}

var _ = Describe("Platform API", func() {
	var gitRepo mocks.GitStorageRepoMock
	var containerRegistryRepo *mocks.ContainerRegistryMock
	var server *httptest.Server
	var crImagesPath string
	var crTagsPath string
	var crTagsPath2 string
	var c http.Client
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
		containerRegistryRepo.StubAndReturnTags2([]containerregistry.ImageTag{{Name: "label1", LastModified: now}})

		logContext := logrus.StandardLogger()
		//k8sClient, k8sConfig := platformK8s.InitKubernetesClient()
		k8sClient := fake.NewSimpleClientset()
		k8sConfig := &rest.Config{}

		k8sPlatformRepoMock := &mocks.K8sPlatformRepoMock{}
		k8sPlatformRepoMock.StubGetSecretAndReturn(&corev1.Secret{})
		k8sRepoV2 := k8s.NewRepo(k8sClient, logContext.WithField("context", "k8s-repo-v2"))

		srv := NewServer(logContext, gitRepo, k8sClient, k8sPlatformRepoMock, k8sRepoV2, k8sConfig, containerRegistryRepo)
		server = httptest.NewServer(srv.Handler)
		crImagesPath = fmt.Sprintf("%s/application/12321/containerregistry/images", server.URL)
		crTagsPath = fmt.Sprintf("%s/application/12321/containerregistry/tags/helloworld", server.URL)
		crTagsPath2 = fmt.Sprintf("%s/application/12321/containerregistry/tags2/helloworld", server.URL)

		c = http.Client{}
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("When fetch images from Container Registry", func() {

		It("should return 403 Forbidden when x-shared-secret header is not set", func() {
			request := createRequest(crImagesPath, "")
			response, _ := c.Do(request)

			Expect(response).ToNot(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("should rertun 403 Forbidden when x-shared-secret header is invalid", func() {
			request := createRequest(crImagesPath, "invalid header")
			response, _ := c.Do(request)

			Expect(response).ToNot(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("should include message in the response when x-shared-secret header is invalid", func() {
			request := createRequest(crImagesPath, "invalid header")
			response, _ := c.Do(request)

			r, _ := ioutil.ReadAll(response.Body)
			var jsonData map[string]interface{}
			json.Unmarshal(r, &jsonData)
			Expect(jsonData["message"]).To(Equal("Shared secret is wrong"))
		})

		It("should return container registry URL from the customer's config in our db (git storage)", func() {
			request := createRequest(crImagesPath, "johnc")
			response, _ := c.Do(request)

			Expect(response).ToNot(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			r, _ := ioutil.ReadAll(response.Body)
			var jsonData map[string]interface{}
			json.Unmarshal(r, &jsonData)
			Expect(jsonData["url"]).To(Equal("test1.azurecr.io"))
		})

		It("should return the images", func() {
			request := createRequest(crImagesPath, "johnc")
			response, _ := c.Do(request)

			Expect(response).ToNot(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			r, _ := ioutil.ReadAll(response.Body)
			var jsonData map[string]interface{}
			json.Unmarshal(r, &jsonData)
			Expect(jsonData["images"]).To(Equal([]interface{}{"helloworld"}))
		})

	})

	Describe("When fetch tags for the images", func() {
		It("should return the tags", func() {
			request := createRequest(crTagsPath, "johnc")
			response, _ := c.Do(request)

			Expect(response).ToNot(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			r, _ := ioutil.ReadAll(response.Body)
			var jsonData map[string]interface{}
			json.Unmarshal(r, &jsonData)
			Expect(jsonData["tags"]).To(Equal([]interface{}{"latest", "v1"}))
		})

	})

	Describe("When fetch tags v2 for the images", func() {
		It("should return the tags with last modified date", func() {
			request := createRequest(crTagsPath2, "johnc")
			response, _ := c.Do(request)

			Expect(response).ToNot(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			r, _ := ioutil.ReadAll(response.Body)
			var jsonData map[string]interface{}
			json.Unmarshal(r, &jsonData)
			Expect(jsonData["tags"]).To(Equal(
				[]interface{}{
					map[string]interface{}{"name": "label1", "lastModified": now.Format(time.RFC3339Nano)},
				},
			))
		})

		It("should rertun 403 Forbidden when x-shared-secret header is invalid", func() {
			request := createRequest(crTagsPath2, "invalid header")
			response, _ := c.Do(request)

			Expect(response).ToNot(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusForbidden))
		})
	})

})
