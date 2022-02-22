package git

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/sirupsen/logrus"
)

func (s *GitStorage) GetMicroserviceDirectory(customerID string, applicationID string, environment string) string {
	return filepath.Join(s.GetRoot(), customerID, applicationID, strings.ToLower(environment))
}

func (s *GitStorage) DeleteMicroservice(customerID string, applicationID string, environment string, microserviceID string) error {
	logContext := s.logContext.WithFields(logrus.Fields{
		"method":          "DeleteMicroservice",
		"customer_id":     customerID,
		"application_id":  applicationID,
		"environment":     environment,
		"microservice_id": microserviceID,
	})

	if err := s.Pull(); err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Pull")
		return err
	}

	dir := s.GetMicroserviceDirectory(customerID, applicationID, environment)
	filename := filepath.Join(dir, fmt.Sprintf("ms_%s.json", microserviceID))
	err := os.Remove(filename)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"filename": filename,
			"error":    err,
		}).Error("Remove")
		return err
	}

	err = s.CommitPathAndPush(filename, fmt.Sprintf("deleted microservice %s", microserviceID))

	if err != nil {
		return err
	}

	return nil
}

func (s *GitStorage) SaveMicroservice(customerID string, applicationID string, environment string, microserviceID string, data interface{}) error {
	logContext := s.logContext.WithFields(logrus.Fields{
		"method":          "SaveMicroservice",
		"customer_id":     customerID,
		"application_id":  applicationID,
		"environment":     environment,
		"microservice_id": microserviceID,
	})

	if err := s.Pull(); err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Pull")
		return err
	}

	dir := s.GetMicroserviceDirectory(customerID, applicationID, environment)
	filename := filepath.Join(dir, fmt.Sprintf("ms_%s.json", microserviceID))
	err := s.writeToDisk(filename, data)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"filename": filename,
			"error":    err,
		}).Error("writeFile")
		return err
	}

	err = s.CommitPathAndPush(filename, fmt.Sprintf("saved microservice %s", microserviceID))

	if err != nil {
		return err
	}

	return nil
}

func (s *GitStorage) GetMicroservice(customerID string, applicationID string, environment string, microserviceID string) ([]byte, error) {
	dir := s.GetMicroserviceDirectory(customerID, applicationID, environment)
	filename := filepath.Join(dir, fmt.Sprintf("ms_%s.json", microserviceID))
	return ioutil.ReadFile(filename)
}

func (s *GitStorage) GetMicroservices(customerID string, applicationID string) ([]platform.HttpMicroserviceBase, error) {
	files := []string{}

	// TODO change
	rootDirectory := s.GetApplicationDirectory(customerID, applicationID)
	// TODO change to fs when gone to 1.16
	err := filepath.Walk(rootDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		// TODO come back to this
		//if info.IsDir() {
		//	fmt.Printf("skipping a dir without errors: %+v \n", info.Name())
		//	return filepath.SkipDir
		//}

		if !strings.HasSuffix(info.Name(), ".json") {
			return nil
		}

		if !strings.HasPrefix(info.Name(), "ms_") {
			return nil
		}

		files = append(files, path)
		return nil
	})

	services := make([]platform.HttpMicroserviceBase, 0)

	if err != nil {
		return services, err
	}

	for _, filename := range files {
		var service platform.HttpMicroserviceBase
		b, _ := ioutil.ReadFile(filename)
		json.Unmarshal(b, &service)
		services = append(services, service)
	}

	return services, nil
}
