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
	"github.com/thoas/go-funk"
)

func (s *GitStorage) GetApplicationDirectory(tenantID string, applicationID string) string {
	return filepath.Join(s.Directory, tenantID, applicationID)
}

func (s *GitStorage) SaveApplication(application platform.HttpResponseApplication) error {
	applicationID := application.ID
	tenantID := application.TenantID
	data, _ := json.MarshalIndent(application, "", " ")

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

	filename := filepath.Join(dir, "application.json")
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
	filename := filepath.Join(dir, "application.json")
	b, err := ioutil.ReadFile(filename)

	var application platform.HttpResponseApplication
	if err != nil {
		return application, err
	}

	err = json.Unmarshal(b, &application)
	if err != nil {
		return application, err
	}

	studioConfig, err := s.GetStudioConfig(tenantID)
	if err != nil {
		return application, err
	}

	// Sprinkle in if automation enabled
	// I wonder if this should be in each applicaiton
	application.Environments = funk.Map(application.Environments, func(e platform.HttpInputEnvironment) platform.HttpInputEnvironment {
		e.AutomationEnabled = s.CheckAutomationEnabledViaCustomer(studioConfig, e.ApplicationID, e.Name)
		return e
	}).([]platform.HttpInputEnvironment)
	return application, nil
}

func (s *GitStorage) GetApplications(tenantID string) ([]platform.HttpResponseApplication, error) {
	applicationIDs := []string{}

	// TODO change to fs when gone to 1.16
	err := filepath.Walk(s.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if info.Name() != "application.json" {
			return nil
		}

		dir := filepath.Dir(path)
		parentDir := filepath.Base(dir)
		applicationID := parentDir

		applicationIDs = append(applicationIDs, applicationID)
		return nil
	})
	applications := make([]platform.HttpResponseApplication, 0)

	if err != nil {
		return applications, err
	}

	for _, applicationID := range applicationIDs {
		application, err := s.GetApplication(tenantID, applicationID)
		if err != nil {
			fmt.Println(err)
			continue
		}
		applications = append(applications, application)
	}

	return applications, nil
}
