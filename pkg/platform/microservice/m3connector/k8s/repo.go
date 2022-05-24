package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/dolittle/platform-api/pkg/platform/microservice/m3connector"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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

// UpsertKafkaFiles will upsert the <env>-kafka-files with the credentials and kafkafiles config
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
	logContext.Info("upserting the kafka files")

	k8sNamespace, err := r.k8sClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return err
	}

	bytesConfig, err := json.MarshalIndent(kafkaFiles.Config, "", "  ")
	if err != nil {
		return err
	}

	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: k8sNamespace.GetAnnotations(),
			Labels:      k8sNamespace.GetLabels(),
		},
		Data: map[string]string{
			"config.json":     string(bytesConfig),
			"accessKey.pem":   kafkaFiles.AccessKey,
			"certificate.pem": kafkaFiles.Certificate,
			"ca.pem":          kafkaFiles.CertificateAuthority,
		},
	}

	logContext.Debug("getting configmap")
	if _, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}

		logContext.Debug("no configmap found, creating a new one")
		if _, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Create(ctx, configMap, metav1.CreateOptions{}); err != nil {
			return err
		}
	} else {
		logContext.Debug("found the config map, updating it")
		_, err = r.k8sClient.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}
