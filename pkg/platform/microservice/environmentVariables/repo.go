package environmentVariables

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type EnvironmentVariablesRepo interface {
	GetEnvironmentVariables(applicationID string, environment string, microserviceID string) ([]platform.StudioEnvironmentVariable, error)
	UpdateEnvironmentVariables(applicationID string, environment string, microserviceID string, data []platform.StudioEnvironmentVariable) error
}

type k8sRepo struct {
	k8sDolittleRepo platform.K8sRepo
	k8sClient       kubernetes.Interface
	logContext      logrus.FieldLogger
}

func NewEnvironmentVariablesK8sRepo(k8sDolittleRepo platform.K8sRepo, k8sClient kubernetes.Interface, logContext logrus.FieldLogger) k8sRepo {
	return k8sRepo{
		k8sDolittleRepo: k8sDolittleRepo,
		k8sClient:       k8sClient,
		logContext:      logContext,
	}
}

func (r k8sRepo) GetEnvironmentVariables(applicationID string, environment string, microserviceID string) ([]platform.StudioEnvironmentVariable, error) {
	// Get name of microservice
	// Get configmap
	// Get secret
	// Build data
	emptyData := make([]platform.StudioEnvironmentVariable, 0)
	data := make([]platform.StudioEnvironmentVariable, 0)
	// TODO this should use environment in GetMicroserviceName
	name, err := r.k8sDolittleRepo.GetMicroserviceName(applicationID, microserviceID)
	if err != nil {
		return data, errors.New("unable to find microservice")
	}

	configmapName := platform.GetMicroserviceEnvironmentVariableConfigmapName(name)

	configMap, err := r.k8sDolittleRepo.GetConfigMap(applicationID, configmapName)
	if err != nil {
		fmt.Println(configmapName)
		fmt.Println(err)
		return data, errors.New("unable to load data from configmap")
	}

	secretName := platform.GetMicroserviceEnvironmentVariableSecretName(name)

	secret, err := r.k8sDolittleRepo.GetSecret(r.logContext, applicationID, secretName)
	if err != nil {
		return data, errors.New("unable to load data from configmap")
	}

	for name, value := range configMap.Data {
		data = append(data, platform.StudioEnvironmentVariable{
			Name:     name,
			Value:    value,
			IsSecret: false,
		})
	}

	for name, value := range secret.Data {
		decodedValue, err := base64.StdEncoding.DecodeString(string(value))
		if err != nil {
			return emptyData, errors.New("bad data")
		}

		data = append(data, platform.StudioEnvironmentVariable{
			Name:     name,
			Value:    string(decodedValue),
			IsSecret: true,
		})
	}

	return data, nil
}

func (r k8sRepo) UpdateEnvironmentVariables(applicationID string, environment string, microserviceID string, data []platform.StudioEnvironmentVariable) error {
	// TODO check for empty name
	// TODO check for empty value
	// TODO check for isSecret
	err := errors.New("bad data")

	for _, item := range data {
		if item.Name == "" {
			return err
		}

		if strings.TrimSpace(item.Name) != item.Name {
			return err
		}

		if item.Value == "" {
			return err
		}

		if strings.TrimSpace(item.Value) != item.Value {
			return err
		}
	}

	// Get name of microservice
	name, err := r.k8sDolittleRepo.GetMicroserviceName(applicationID, microserviceID)
	if err != nil {
		return errors.New("unable to find microservice")
	}

	configmapName := platform.GetMicroserviceEnvironmentVariableConfigmapName(name)
	configMap, err := r.k8sDolittleRepo.GetConfigMap(applicationID, configmapName)
	if err != nil {
		return errors.New("unable to load data from configmap")
	}

	secretName := platform.GetMicroserviceEnvironmentVariableSecretName(name)
	secret, err := r.k8sDolittleRepo.GetSecret(r.logContext, applicationID, secretName)
	if err != nil {
		return errors.New("unable to load data from configmap")
	}

	// TODO would be nice to use a resource (application-namespace branch)
	//r.k8sDolittleRepo.WriteConfigMap
	// Update data
	configMap.Data = make(map[string]string)

	for _, item := range data {
		if item.IsSecret {
			continue
		}
		configMap.Data[item.Name] = item.Value
	}

	secret.StringData = make(map[string]string)

	for _, item := range data {
		if !item.IsSecret {
			continue
		}
		secret.StringData[item.Name] = item.Value
	}

	// Write configmap and secret
	_, err = r.k8sDolittleRepo.WriteConfigMap(configMap)
	if err != nil {
		return errors.New("failed to update configmap")
	}

	_, err = r.k8sDolittleRepo.WriteSecret(secret)
	if err != nil {
		return errors.New("failed to update secret")
	}
	return nil
}
