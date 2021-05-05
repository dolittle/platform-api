package microservice

import (
	"errors"
	"fmt"

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

func (s *k8sStorage) Write(info HttpInputDolittle, data []byte) error {
	fmt.Printf("Write %s.json to file", info.MicroserviceID)
	return errors.New("TODO")
}

func (s *k8sStorage) Read(info HttpInputDolittle) ([]byte, error) {
	fmt.Printf("%v", info)
	return []byte(`{"TODO":"TODO"}`), nil
}

func (s *k8sStorage) GetAll(tenantID string, applicationID string) ([]HttpMicroserviceBase, error) {
	services := make([]HttpMicroserviceBase, 0)
	return services, errors.New("TODO")
}
