package backup

import (
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type service struct {
	gitRepo         storage.Repo
	k8sDolittleRepo platform.K8sRepo
	k8sClient       *kubernetes.Clientset
	logContext      logrus.FieldLogger
}

type AzureStorageInfo struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type HTTPDownloadLogsLatestResponse struct {
	Application platform.ShortInfo `json:"application"`
	Environment string             `json:"environment"`
	Files       []string           `json:"files"`
}

type HTTPDownloadLogsInput struct {
	ApplicationID string `json:"applicationId"`
	Environment   string `json:"environment"`
	FilePath      string `json:"file_path"`
}

type HTTPDownloadLogsLinkResponse struct {
	Application platform.ShortInfo `json:"application"`
	Url         string             `json:"url"`
	Expires     string             `json:"expire"`
}
