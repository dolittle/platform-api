package api

import (
	"net/http/httptest"

	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/storage/git"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var _ = Describe("foo", func() {

	It("should do something", func() {
		viper.Set("tools.server.gitRepo.branch", "foo")
		viper.Set("tools.server.gitRepo.directory", "/tmp/dolittle_operation")
		viper.Set("tools.server.gitRepo.url", "git@github.com:dolittle-platform/Operations")
		viper.Set("tools.server.gitRepo.sshKey", "does/not/exist")
		viper.Set("tools.server.kubernetes.externalClusterHost", "external-host")

		logContext := logrus.StandardLogger()
		gitStorageCfg := git.GitStorageConfig{}
		k8sClient, k8sConfig := platformK8s.InitKubernetesClient()

		srv := NewServer(logContext, gitStorageCfg, k8sClient, k8sConfig)
		Expect(1).To(Equal(2))
		s := httptest.NewServer(srv.Handler)
		defer s.Close()

		Expect(s.URL).To(Equal("foo"))

	})

})
