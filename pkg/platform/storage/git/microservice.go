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
)

func (s *GitStorage) GetMicroserviceDirectory(tenantID string, applicationID string, environment string) string {
	return fmt.Sprintf("%s/%s/%s/%s", s.Directory, tenantID, applicationID, strings.ToLower(environment))
}

func (s *GitStorage) SaveMicroservice(tenantID string, applicationID string, environment string, microserviceID string, data []byte) error {
	w, err := s.Repo.Worktree()
	if err != nil {
		return err
	}

	// TODO actually build structure
	// `{tenantID}/{applicationID}/{environment}/{microserviceID}.json`
	dir := s.GetMicroserviceDirectory(tenantID, applicationID, environment)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s/ms_%s.json", dir, microserviceID)
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		fmt.Println("writeFile")
		return err
	}

	// Adds the new file to the staging area.
	// Need to remove the prefix
	err = w.AddWithOptions(&git.AddOptions{
		Path: strings.TrimPrefix(filename, s.Directory+"/"),
	})

	if err != nil {
		fmt.Println("w.Add")
		return err
	}

	_, err = w.Status()
	if err != nil {
		fmt.Println("w.Status")
		return err
	}

	err = s.CommitAndPush(w, "upsert microservice")

	if err != nil {
		return err
	}

	return nil
}

func (s *GitStorage) GetMicroservice(tenantID string, applicationID string, environment string, microserviceID string) ([]byte, error) {
	dir := s.GetMicroserviceDirectory(tenantID, applicationID, environment)
	filename := fmt.Sprintf("%s/ms_%s.json", dir, microserviceID)
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
