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

func (s *GitStorage) GetMicroserviceDirectory(tenantID string, applicationID string, environment string) string {
	return filepath.Join(s.GetRoot(), tenantID, applicationID, strings.ToLower(environment))
}

func (s *GitStorage) DeleteMicroservice(tenantID string, applicationID string, environment string, microserviceID string) error {
	logContext := s.logContext.WithFields(logrus.Fields{
		"method":         "DeleteMicroservice",
		"tenantID":       tenantID,
		"applicationID":  applicationID,
		"environment":    environment,
		"microserviceID": microserviceID,
	})

	if err := s.Pull(); err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Pull")
		return err
	}

	dir := s.GetMicroserviceDirectory(tenantID, applicationID, environment)
	filename := filepath.Join(dir, fmt.Sprintf("ms_%s.json", microserviceID))
	err := os.Remove(filename)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"filename": filename,
			"error":    err,
		}).Error("Remove")
		return err
	}

	// Need to remove the prefix
	path := strings.TrimPrefix(filename, s.config.RepoRoot+string(os.PathSeparator))

	err = s.CommitPathAndPush(path, fmt.Sprintf("deleted microservice %s", microserviceID))

	if err != nil {
		return err
	}

	return nil
}

func (s *GitStorage) SaveMicroservice(tenantID string, applicationID string, environment string, microserviceID string, data interface{}) error {
	logContext := s.logContext.WithFields(logrus.Fields{
		"method":         "SaveMicroservice",
		"tenantID":       tenantID,
		"applicationID":  applicationID,
		"environment":    environment,
		"microserviceID": microserviceID,
	})
	storageBytes, _ := json.MarshalIndent(data, "", "  ")
	if err := s.Pull(); err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Pull")
		return err
	}

	dir := s.GetMicroserviceDirectory(tenantID, applicationID, environment)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	filename := filepath.Join(dir, fmt.Sprintf("ms_%s.json", microserviceID))
	err = ioutil.WriteFile(filename, storageBytes, 0644)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"filename": filename,
			"error":    err,
		}).Error("writeFile")
		return err
	}

	// Need to remove the prefix
	path := strings.TrimPrefix(filename, s.config.RepoRoot+string(os.PathSeparator))

	err = s.CommitPathAndPush(path, fmt.Sprintf("saved microservice %s", microserviceID))

	if err != nil {
		return err
	}

	return nil
}

func (s *GitStorage) GetMicroservice(tenantID string, applicationID string, environment string, microserviceID string) ([]byte, error) {
	dir := s.GetMicroserviceDirectory(tenantID, applicationID, environment)
	filename := filepath.Join(dir, fmt.Sprintf("ms_%s.json", microserviceID))
	return ioutil.ReadFile(filename)
}

func (s *GitStorage) GetMicroservices(tenantID string, applicationID string) ([]platform.HttpMicroserviceBase, error) {
	files := []string{}

	// TODO change
	rootDirectory := s.GetApplicationDirectory(tenantID, applicationID)
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
