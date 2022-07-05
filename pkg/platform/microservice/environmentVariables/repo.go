package environmentVariables

import (
	"errors"
	"strings"

	"github.com/dolittle/platform-api/pkg/platform"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"k8s.io/client-go/kubernetes"
)

type EnvironmentVariablesRepo interface {
	GetEnvironmentVariables(applicationID string, environment string, microserviceID string) ([]platform.StudioEnvironmentVariable, error)
	UpdateEnvironmentVariables(applicationID string, environment string, microserviceID string, data []platform.StudioEnvironmentVariable) error
}

type k8sRepo struct {
	k8sDolittleRepo platformK8s.K8sPlatformRepo
	k8sClient       kubernetes.Interface
	logContext      logrus.FieldLogger
}

func NewEnvironmentVariablesK8sRepo(k8sDolittleRepo platformK8s.K8sPlatformRepo, k8sClient kubernetes.Interface, logContext logrus.FieldLogger) k8sRepo {
	return k8sRepo{
		k8sDolittleRepo: k8sDolittleRepo,
		k8sClient:       k8sClient,
		logContext:      logContext,
	}
}

func (r k8sRepo) GetEnvironmentVariables(applicationID string, environment string, microserviceID string) ([]platform.StudioEnvironmentVariable, error) {
	data := make([]platform.StudioEnvironmentVariable, 0)
	name, err := r.k8sDolittleRepo.GetMicroserviceName(applicationID, environment, microserviceID)
	if err != nil {
		return data, errors.New("unable to find microservice")
	}

	configmapName := platformK8s.GetMicroserviceEnvironmentVariableConfigmapName(name)

	configMap, err := r.k8sDolittleRepo.GetConfigMap(applicationID, configmapName)
	if err != nil {
		return data, errors.New("unable to load data from configmap")
	}

	secretName := platformK8s.GetMicroserviceEnvironmentVariableSecretName(name)

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

	// When using StringData, it does not appear I need to handle this
	for name, value := range secret.Data {
		data = append(data, platform.StudioEnvironmentVariable{
			Name:     name,
			Value:    string(value),
			IsSecret: true,
		})
	}

	return data, nil
}

func (r k8sRepo) UpdateEnvironmentVariables(applicationID string, environment string, microserviceID string, data []platform.StudioEnvironmentVariable) error {
	err := errors.New("bad data")
	uniqueNames := make([]string, 0)
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

		// Check for duplicate keys
		if funk.ContainsString(uniqueNames, item.Name) {
			return err
		}

		uniqueNames = append(uniqueNames, item.Name)
	}

	// Get name of microservice
	name, err := r.k8sDolittleRepo.GetMicroserviceName(applicationID, environment, microserviceID)
	if err != nil {
		return errors.New("unable to find microservice")
	}

	configmapName := platformK8s.GetMicroserviceEnvironmentVariableConfigmapName(name)
	configMap, err := r.k8sDolittleRepo.GetConfigMap(applicationID, configmapName)
	if err != nil {
		return errors.New("unable to load data from configmap")
	}

	secretName := platformK8s.GetMicroserviceEnvironmentVariableSecretName(name)
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

	// Because I am overriding all, this is required to make sure
	// Current data is cleared, as I suspect StringData has some magic under the hood on
	// saving and on reading
	secret.Data = make(map[string][]byte)
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
