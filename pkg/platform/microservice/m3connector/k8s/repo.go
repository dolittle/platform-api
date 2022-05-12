package k8s

import (
	"github.com/dolittle/platform-api/pkg/platform/microservice/m3connector"
	"k8s.io/client-go/kubernetes"
)

type k8sRepo struct {
	k8sClient    kubernetes.Interface
	isProduction bool
}

func NewM3ConnectorRepo(k8sClient kubernetes.Interface, isProduction bool) m3connector.K8sRepo {
	return &k8sRepo{
		k8sClient:    k8sClient,
		isProduction: isProduction,
	}
}

func (r *k8sRepo) UpsertKafkaFiles(applicationID, environment string, kafkaFiles m3connector.KafkaFiles) error {

	return nil
}
