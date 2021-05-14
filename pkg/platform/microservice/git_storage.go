package microservice

import (
	"encoding/json"
	"fmt"
	"strings"

	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type gitRepo struct {
	storage *platform.GitStorage
}

func NewGitRepo(storage *platform.GitStorage) *gitRepo {
	return &gitRepo{
		storage: storage,
	}
}

func (s *gitRepo) getDirectory(tenantID string, applicationID string, environment string) string {
	return fmt.Sprintf("%s/%s/%s/%s", s.storage.Directory, tenantID, applicationID, strings.ToLower(environment))
}

func (s *gitRepo) Write(tenantID string, applicationID string, environment string, microserviceID string, data []byte) error {
	w, err := s.storage.Repo.Worktree()
	if err != nil {
		return err
	}

	// TODO actually build structure
	// `{tenantID}/{applicationID}/{environment}/{microserviceID}.json`
	dir := s.getDirectory(tenantID, applicationID, environment)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s/%s.json", dir, microserviceID)
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		fmt.Println("writeFile")
		return err
	}

	// Adds the new file to the staging area.
	// Need to remove the prefix
	err = w.AddWithOptions(&git.AddOptions{
		Path: strings.TrimPrefix(filename, s.storage.Directory+"/"),
	})

	if err != nil {
		fmt.Println("w.Add")
		return err
	}

	status, err := w.Status()
	if err != nil {
		fmt.Println("w.Status")
		return err
	}

	fmt.Println(status)

	commit, err := w.Commit("example go-git commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  time.Now(),
		},
	})

	if err != nil {
		return err
	}

	// Prints the current HEAD to verify that all worked well.
	_, err = s.storage.Repo.CommitObject(commit)
	return err
}

func (s *gitRepo) Read(tenantID string, applicationID string, environment string, microserviceID string) ([]byte, error) {
	dir := s.getDirectory(tenantID, applicationID, environment)
	filename := fmt.Sprintf("%s/%s.json", dir, microserviceID)
	return ioutil.ReadFile(filename)
}

func (s *gitRepo) GetAll(tenantID string, applicationID string) ([]HttpMicroserviceBase, error) {
	files := []string{}

	// TODO change
	rootDirectory := s.storage.GetApplicationDirectory(tenantID, applicationID)
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

		files = append(files, path)
		return nil
	})

	services := make([]HttpMicroserviceBase, 0)

	if err != nil {
		return services, err
	}

	for _, filename := range files {
		var service HttpMicroserviceBase
		b, _ := ioutil.ReadFile(filename)
		json.Unmarshal(b, &service)
		services = append(services, service)
	}

	return services, nil
}
