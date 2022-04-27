package configFiles

import (
	"errors"

	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/sirupsen/logrus"
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
	// err := errors.New("bad data")
	// uniqueNames := make([]string, 0)
	// for _, item := range data {
	// 	if item.Name == "" {
	// 		return err
	// 	}

	// 	if strings.TrimSpace(item.Name) != item.Name {
	// 		return err
	// 	}

	// 	if item.Value == "" {
	// 		return err
	// 	}

	// 	if strings.TrimSpace(item.Value) != item.Value {
	// 		return err
	// 	}

	// 	// Check for duplicate keys
	// 	if funk.ContainsString(uniqueNames, item.Name) {
	// 		return err
	// 	}

	// 	uniqueNames = append(uniqueNames, item.Name)
	// }

	// // Get name of microservice
	// name, err := r.k8sDolittleRepo.GetMicroserviceName(applicationID, environment, microserviceID)
	// if err != nil {
	// 	return errors.New("unable to find microservice")
	// }

	// configmapName := platformK8s.GetMicroserviceConfigFilesConfigmapName(name)
	// configMap, err := r.k8sDolittleRepo.GetConfigMap(applicationID, configmapName)
	// if err != nil {
	// 	return errors.New("unable to load data from configmap")
	// }

	// // TODO would be nice to use a resource (application-namespace branch)
	// //r.k8sDolittleRepo.WriteConfigMap
	// // Update data
	// configMap.Data = make(map[string]string)

	// for _, item := range data {
	// 	configMap.Data[item.Name] = item.Value
	// }

	// // Write configmap and secret
	// _, err = r.k8sDolittleRepo.WriteConfigMap(configMap)
	// if err != nil {
	// 	return errors.New("failed to update configmap")
	// }

	return nil
}
