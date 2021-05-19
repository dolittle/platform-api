package microservice

import (
	"errors"
	"fmt"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"k8s.io/client-go/kubernetes"
)

type k8sStorage struct {
	k8sClient *kubernetes.Clientset
}

func NewK8sStorage(k8sClient *kubernetes.Clientset) *k8sStorage {
	return &k8sStorage{
		k8sClient: k8sClient,
	}
}

func (s *k8sStorage) Write(info platform.HttpInputDolittle, data []byte) error {
	fmt.Printf("Write %s.json to file", info.MicroserviceID)
	return errors.New("TODO")
}

func (s *k8sStorage) Read(info platform.HttpInputDolittle) ([]byte, error) {
	fmt.Printf("%v", info)
	return []byte(`{"TODO":"TODO"}`), nil
}

func (s *k8sStorage) GetAll(tenantID string, applicationID string) ([]platform.HttpMicroserviceBase, error) {
	services := make([]platform.HttpMicroserviceBase, 0)
	return services, errors.New("TODO")
}
