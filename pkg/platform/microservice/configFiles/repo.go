package configFiles

import (
	"errors"
	"fmt"

	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type ConfigFilesRepo interface {
	GetConfigFilesNamesList(applicationID string, environment string, microserviceID string) ([]string, error)
	AddEntryToConfigFiles(applicationID string, environment string, microserviceID string, data platform.StudioConfigFile) error
	RemoveEntryFromConfigFiles(applicationID string, environment string, microserviceID string, key string) error
}

type k8sRepo struct {
	k8sDolittleRepo platformK8s.K8sRepo
	k8sClient       kubernetes.Interface
	logContext      logrus.FieldLogger
}

func NewConfigFilesK8sRepo(k8sDolittleRepo platformK8s.K8sRepo, k8sClient kubernetes.Interface, logContext logrus.FieldLogger) k8sRepo {
	return k8sRepo{
		k8sDolittleRepo: k8sDolittleRepo,
		k8sClient:       k8sClient,
		logContext:      logContext,
	}
}

func (r k8sRepo) GetConfigFilesNamesList(applicationID string, environment string, microserviceID string) ([]string, error) {
	data := []string{}

	logContext := r.logContext.WithFields(logrus.Fields{
		"method": "GetConfigFilesNamesList",
		"application_id": applicationID,
		"microservice_id": microserviceID,
		"environment": environment,
	})

	name, err := r.k8sDolittleRepo.GetMicroserviceName(applicationID, environment, microserviceID)
	if err != nil {
		logContext.WithField("error", err).Error("unable to find microservice")
		return data, err
	}

	configmapName := platformK8s.GetMicroserviceConfigFilesConfigmapName(name)

	configMap, err := r.k8sDolittleRepo.GetConfigMap(applicationID, configmapName)
	if err != nil {
		logContext.WithField("error", err).Error("unable to load data from configmap")
		return data, err
	}


	for name := range configMap.BinaryData {
		data = append(data, name)
	}

	return data, nil
}

func (r k8sRepo) AddEntryToConfigFiles(applicationID string, environment string, microserviceID string, data platform.StudioConfigFile) error {

	// Get name of microservice
	name, err := r.k8sDolittleRepo.GetMicroserviceName(applicationID, environment, microserviceID)
	if err != nil {
		return errors.New("unable to find microservice")
	}

	configmapName := platformK8s.GetMicroserviceConfigFilesConfigmapName(name)
	configMap, err := r.k8sDolittleRepo.GetConfigMap(applicationID, configmapName)

	if len(configMap.Data) == 0 {
		configMap.Data = map[string]string{}
	}

	if len(configMap.BinaryData) == 0 {
		configMap.BinaryData = map[string][]byte{}
	}

	if err != nil {
		return errors.New("unable to load data from configmap")
	}

	// TODO would be nice to use a resource (application-namespace branch)
	//r.k8sDolittleRepo.WriteConfigMap
	// Update data

	configMap.BinaryData[data.Name] = data.BinaryData

	// Write configmap and secret
	_, err = r.k8sDolittleRepo.WriteConfigMap(configMap)
	if err != nil {
		fmt.Println()
		return errors.New("failed to update configmap: " + err.Error())
	}

	return nil
}

func (r k8sRepo) RemoveEntryFromConfigFiles(applicationID string, environment string, microserviceID string, key string) error {

	logContext := r.logContext.WithFields(logrus.Fields{
		"method": "GetConfigFilesNamesList",
		"application_id": applicationID,
		"microservice_id": microserviceID,
		"environment": environment,
	})

	// Get name of microservice
	name, err := r.k8sDolittleRepo.GetMicroserviceName(applicationID, environment, microserviceID)
	if err != nil {
		logContext.WithField("error", err).Error("unable to find microservice")
		return err
	}

	configmapName := platformK8s.GetMicroserviceConfigFilesConfigmapName(name)
	configMap, err := r.k8sDolittleRepo.GetConfigMap(applicationID, configmapName)

	if err != nil {
		logContext.WithField("error", err).Error("unable to load data from configmap")
		return err
	}

	if len(configMap.BinaryData) == 0 {
		logContext.WithField("error", err).Error("no entries in configmap")
		return err
	}

	delete(configMap.BinaryData, key)

	// Write configmap and secret
	_, err = r.k8sDolittleRepo.WriteConfigMap(configMap)

	if err != nil {
		logContext.WithField("error", err).Error("failed to update configmap")
		return err;
	}

	return nil
}