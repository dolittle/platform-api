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

func (s *GitStorage) GetApplicationDirectory(tenantID string, applicationID string) string {
	return fmt.Sprintf("%s/%s/%s", s.Directory, tenantID, applicationID)
}

func (s *GitStorage) SaveApplication(application platform.HttpResponseApplication) error {
	applicationID := application.ID
	tenantID := application.TenantID
	data, _ := json.Marshal(application)

	w, err := s.Repo.Worktree()
	if err != nil {
		return err
	}

	// TODO actually build structure
	dir := s.GetApplicationDirectory(tenantID, applicationID)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s/application.json", dir)
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

	err = s.CommitAndPush(w, "upsert application")

	if err != nil {
		return err
	}

	return nil
}

func (s *GitStorage) GetApplication(tenantID string, applicationID string) (platform.HttpResponseApplication, error) {
	dir := s.GetApplicationDirectory(tenantID, applicationID)
	filename := fmt.Sprintf("%s/application.json", dir)
	b, err := ioutil.ReadFile(filename)

	var application platform.HttpResponseApplication
	if err != nil {
		return application, err
	}

	err = json.Unmarshal(b, &application)
	if err != nil {
		return application, err
	}
	return application, nil
}

func (s *GitStorage) GetApplications(tenantID string) ([]platform.HttpResponseApplication, error) {
	files := []string{}

	// TODO change
	rootDirectory := s.Directory + "/"
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

		if info.Name() != "application.json" {
			return nil
		}

		files = append(files, path)
		return nil
	})

	applications := make([]platform.HttpResponseApplication, 0)

	if err != nil {
		return applications, err
	}

	for _, filename := range files {
		var application platform.HttpResponseApplication
		b, _ := ioutil.ReadFile(filename)
		json.Unmarshal(b, &application)
		applications = append(applications, application)
	}

	return applications, nil
}
