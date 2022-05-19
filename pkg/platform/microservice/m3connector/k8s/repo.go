package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/m3connector"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type k8sRepo struct {
	k8sClient    kubernetes.Interface
	isProduction bool
	logger       logrus.FieldLogger
}

func NewM3ConnectorRepo(k8sClient kubernetes.Interface, isProduction bool, logger *logrus.Logger) m3connector.K8sRepo {
	return &k8sRepo{
		k8sClient:    k8sClient,
		isProduction: isProduction,
		logger:       logger.WithField("repo", "m3-k8s-repo"),
	}
}

func (r *k8sRepo) UpsertKafkaFiles(applicationID, environment string, kafkaFiles m3connector.KafkaFiles) error {
	ctx := context.TODO()
	namespace := platformK8s.GetApplicationNamespace(applicationID)
	name := fmt.Sprintf("%s-kafka-files", environment)
	logContext := r.logger.WithFields(logrus.Fields{
		"application_id": applicationID,
		"namespace":      namespace,
		"configmap_name": name,
		"method":         "UpsertKafkaFiles",
	})
	logContext.Debug("fetcing configmap")

	configMap, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	bytesConfig, err := json.MarshalIndent(kafkaFiles.Config, "", "  ")
	if err != nil {
		return err
	}
	configMap.Data["config.json"] = string(bytesConfig)

	configMap.Data["accessKey.pem"] = kafkaFiles.AccessKey
	configMap.Data["certificate.pem"] = kafkaFiles.Certificate
	configMap.Data["ca.pem"] = kafkaFiles.CertificateAuthority

	_, err = r.k8sClient.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}
