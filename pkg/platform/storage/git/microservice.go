package git

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	git "github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"
)

func (s *GitStorage) GetMicroserviceDirectory(tenantID string, applicationID string, environment string) string {
	return filepath.Join(s.Directory, tenantID, applicationID, strings.ToLower(environment))
}

func (s *GitStorage) DeleteMicroservice(tenantID string, applicationID string, environment string, microserviceID string) error {
	logContext := s.logContext.WithFields(logrus.Fields{
		"method":         "DeleteMicroservice",
		"tenantID":       tenantID,
		"applicationID":  applicationID,
		"environment":    environment,
		"microserviceID": microserviceID,
	})
	w, err := s.Repo.Worktree()
	if err != nil {
		return err
	}
	if err = s.PullWithWorktree(w); err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("PullWithWorktree")
		return err
	}

	dir := s.GetMicroserviceDirectory(tenantID, applicationID, environment)
	filename := filepath.Join(dir, fmt.Sprintf("ms_%s.json", microserviceID))
	err = os.Remove(filename)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"filename": filename,
			"error":    err,
		}).Error("Remove")
		return err
	}

	// Adds the new file to the staging area.
	// Need to remove the prefix
	path := strings.TrimPrefix(filename, s.Directory+"/")
	err = w.AddWithOptions(&git.AddOptions{
		Path: path,
	})

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"addPath": path,
			"error":   err,
		}).Error("Failed to AddWithOptions")
		return err
	}

	_, err = w.Status()
	if err != nil {
		fmt.Println("w.Status")
		return err
	}

	err = s.CommitAndPush(w, "deleted microservice")

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
	w, err := s.Repo.Worktree()
	if err != nil {
		return err
	}

	if err = s.PullWithWorktree(w); err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("PullWithWorktree")
		return err
	}

	// TODO actually build structure
	// `{tenantID}/{applicationID}/{environment}/{microserviceID}.json`
	dir := s.GetMicroserviceDirectory(tenantID, applicationID, environment)
	err = os.MkdirAll(dir, 0755)
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

	// Adds the new file to the staging area.
	// Need to remove the prefix
	path := strings.TrimPrefix(filename, s.Directory+"/")
	err = w.AddWithOptions(&git.AddOptions{
		Path: path,
	})

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"addPath": path,
			"error":   err,
		}).Error("Failed to AddWithOptions")
		return err
	}

	_, err = w.Status()
	if err != nil {
		logContext.Error("Failed to call Status on Worktree")
		return err
	}

	err = s.CommitAndPush(w, "saved microservice")

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
