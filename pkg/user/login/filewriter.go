package login

import (
	"io/ioutil"

	corev1 "k8s.io/api/core/v1"
)

type Repo interface {
	Get() (LoginConfiguration, error)
	Save(data LoginConfiguration) error
}

type fileWriter struct {
	pathToFile string
}

func NewFileWriter(pathToFile string) Repo {

	//pathToFile := filepath.Join(rootDirectory, "Source", "V3", "Kubernetes", "System", "Authentication", "Dev", "users.yml")

	return fileWriter{
		pathToFile: pathToFile,
	}
}

func (w fileWriter) Get() (LoginConfiguration, error) {
	// Source/V3/Kubernetes/System/Authentication/Dev/users.yml

	fileBytes, err := ioutil.ReadFile(w.pathToFile)
	if err != nil {
		panic(err)
	}
	var configMap corev1.ConfigMap
	err = configMap.Unmarshal(fileBytes)
	if err != nil {
		panic(err)
	}

	return LoginConfiguration{}, nil
}

func (w fileWriter) Save(data LoginConfiguration) error {
	return nil
}
