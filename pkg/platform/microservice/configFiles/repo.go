package configFiles

import (
	"errors"

	"strings"

	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"k8s.io/client-go/kubernetes"
)

type ConfigFilesRepo interface {
	GetConfigFile(applicationID string, environment string, microserviceID string) (platform.StudioConfigFile, error)
	UpdateConfigFiles(applicationID string, environment string, microserviceID string, data platform.StudioConfigFile) error
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

func (r k8sRepo) GetConfigFile(applicationID string, environment string, microserviceID string) (platform.StudioConfigFile, error) {
	var data platform.StudioConfigFile

	name, err := r.k8sDolittleRepo.GetMicroserviceName(applicationID, environment, microserviceID)
	if err != nil {
		return data, errors.New("unable to find microservice")
	}

	configmapName := platformK8s.GetMicroserviceConfigFilesConfigmapName(name)

	configMap, err := r.k8sDolittleRepo.GetConfigMap(applicationID, configmapName)
	if err != nil {
		return data, errors.New("unable to load data from configmap")
	}

	if err != nil {
		return data, errors.New("unable to load data from configmap")
	}

	for name, value := range configMap.Data {
		data.Name = name
		data.Value = value
	}

	return data, nil
}

func (r k8sRepo) UpdateConfigFiles(applicationID string, environment string, microserviceID string, data platform.StudioConfigFile) error {

	// Get name of microservice
	name, err := r.k8sDolittleRepo.GetMicroserviceName(applicationID, environment, microserviceID)
	if err != nil {
		return errors.New("unable to find microservice")
	}

	configmapName := platformK8s.GetMicroserviceConfigFilesConfigmapName(name)
	configMap, err := r.k8sDolittleRepo.GetConfigMap(applicationID, configmapName)
	if err != nil {
		return errors.New("unable to load data from configmap")
	}

	// VICTOR TODO: Replace this validaton with correct validation
	uniqueNames := make([]string, 0)

	for name, value := range configMap.Data {
		if name == "" {
			return errors.New("Empty config file name in existing configmap")
		}

		if strings.TrimSpace(name) != name {
			return errors.New("No spaces allowed in config file name in existing configmap")
		}

		if value == "" {
			return errors.New("No empty value allowed in config file in existing configmap")
		}

		if strings.TrimSpace(value) != value {
			return errors.New("No spaces allowed in config file value in config file in existing configmap")
		}

		// Check for duplicate keys
		if funk.ContainsString(uniqueNames, name) {
			return errors.New("No duplicate keys allowed in config file in existing configmap")
		}

		uniqueNames = append(uniqueNames, name)
	}

	// TODO would be nice to use a resource (application-namespace branch)
	//r.k8sDolittleRepo.WriteConfigMap
	// Update data

	configMap.Data[data.Name] = " | \n" + data.Value

	// Write configmap and secret
	_, err = r.k8sDolittleRepo.WriteConfigMap(configMap)
	if err != nil {
		return errors.New("failed to update configmap")
	}

	return nil
}
