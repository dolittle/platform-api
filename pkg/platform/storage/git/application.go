package git

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	git "github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
)

func (s *GitStorage) GetApplicationDirectory(tenantID string, applicationID string) string {
	return filepath.Join(s.Directory, tenantID, applicationID)
}

func (s *GitStorage) SaveApplication(application platform.HttpResponseApplication) error {
	applicationID := application.ID
	tenantID := application.TenantID
	logContext := s.logContext.WithFields(log.Fields{
		"method":        "SaveApplication",
		"customer":      tenantID,
		"applicationID": applicationID,
	})
	data, _ := json.MarshalIndent(application, "", " ")

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
	dir := s.GetApplicationDirectory(tenantID, applicationID)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	filename := filepath.Join(dir, "application.json")
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		logContext.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to write 'application.json'")
		return err
	}

	// Adds the new file to the staging area.
	// Need to remove the prefix
	addPath := strings.TrimPrefix(filename, s.Directory+string(os.PathSeparator))
	err = w.AddWithOptions(&git.AddOptions{
		Path: addPath,
	})
	if err != nil {
		logContext.WithFields(log.Fields{
			"path":  addPath,
			"error": err,
		}).Error("Failed to add path to worktree")
		return err
	}

	_, err = w.Status()
	if err != nil {
		logContext.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to get worktree status")
		return err
	}

	err = s.CommitAndPush(w, "upsert application")
	if err != nil {
		logContext.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to commit and push worktree")
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

	application.Environments = funk.Map(application.Environments, func(e platform.HttpInputEnvironment) platform.HttpInputEnvironment {
		e.AutomationEnabled = s.IsAutomationEnabled(tenantID, e.ApplicationID, e.Name)
		return e
	}).([]platform.HttpInputEnvironment)
	return application, nil
}

func (s *GitStorage) GetApplications(customerID string) ([]platform.HttpResponseApplication, error) {
	applicationIDs, err := s.discoverCustomerApplicationIds(customerID)
	applications := make([]platform.HttpResponseApplication, 0)

	if err != nil {
		return applications, err
	}

	for _, applicationID := range applicationIDs {
		application, err := s.GetApplication(customerID, applicationID)
		if err != nil {
			s.logContext.WithFields(log.Fields{
				"customer":    customerID,
				"application": applicationID,
				"error":       err,
			}).Warning("Skipping application because it failed to load")
			continue
		}
		applications = append(applications, application)
	}

	return applications, nil
}
func (s *GitStorage) discoverCustomerApplicationIds(customerID string) ([]string, error) {
	applicationIDs := []string{}

	// TODO change to fs when gone to 1.16
	err := filepath.Walk(s.GetTenantDirectory(customerID), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			s.logContext.WithFields(log.Fields{
				"customer": customerID,
				"path":     path,
				"error":    err,
			}).Error("prevent panic by handling failure accessing a path")
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

	return applicationIDs, err
}
