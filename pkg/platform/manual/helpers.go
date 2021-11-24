package manual

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/thoas/go-funk"
	"gopkg.in/yaml.v3"
)

type MicroserviceInfo struct {
	ApplicationID  string `json:"applicationId"`
	Environment    string `json:"environment"`
	MicroserviceID string `json:"microserviceId"`
	TenantID       string `json:"tenantId"`
	Namespace      string `json:"namespace"`
}

type PlatformMicroserviceInfo struct {
	Metadata PlatformMicroserviceInfoMetadata `yaml:"metadata"`
}
type PlatformMicroserviceInfoAnnotations struct {
	DolittleIoTenantID       string `yaml:"dolittle.io/tenant-id"`
	DolittleIoApplicationID  string `yaml:"dolittle.io/application-id"`
	DolittleIoMicroserviceID string `yaml:"dolittle.io/microservice-id"`
}
type PlatformMicroserviceInfoLabels struct {
	Tenant       string `yaml:"tenant"`
	Application  string `yaml:"application"`
	Environment  string `yaml:"environment"`
	Microservice string `yaml:"microservice"`
}

type PlatformMicroserviceInfoMetadata struct {
	Annotations PlatformMicroserviceInfoAnnotations `yaml:"annotations"`
	Labels      PlatformMicroserviceInfoLabels      `yaml:"labels"`
	Name        string                              `yaml:"name"`
	Namespace   string                              `yaml:"namespace"`
}

func GetMicroservicePaths(rootDir string) ([]string, error) {
	microservicePaths := []string{}
	err := filepath.Walk(rootDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !strings.Contains(path, "/Source/V3/Kubernetes/Customers/") {
				return nil
			}

			if !funk.ContainsString([]string{"microservice.yml", "microservice-deployment.yml"}, info.Name()) {
				return nil
			}
			microservicePaths = append(microservicePaths, path)
			return nil
		})

	return microservicePaths, err
}

func GetMicroserviceInfo(filePath string) MicroserviceInfo {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return MicroserviceInfo{}
	}
	resources, err := SplitYAML(data)
	if err != nil {
		return MicroserviceInfo{}
	}

	var rawInfo PlatformMicroserviceInfo
	yaml.Unmarshal(resources[0], &rawInfo)

	return MicroserviceInfo{
		MicroserviceID: rawInfo.Metadata.Annotations.DolittleIoMicroserviceID,
		ApplicationID:  rawInfo.Metadata.Annotations.DolittleIoApplicationID,
		TenantID:       rawInfo.Metadata.Annotations.DolittleIoTenantID,
		Environment:    rawInfo.Metadata.Labels.Environment,
		Namespace:      rawInfo.Metadata.Namespace,
	}
}

func SplitYAML(resources []byte) ([][]byte, error) {
	dec := yaml.NewDecoder(bytes.NewReader(resources))

	var res [][]byte
	for {
		var value interface{}
		err := dec.Decode(&value)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		valueBytes, err := yaml.Marshal(value)
		if err != nil {
			return nil, err
		}
		res = append(res, valueBytes)
	}
	return res, nil
}
