package git

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/sirupsen/logrus"
)

func (s *GitStorage) GetApplicationDirectory(tenantID string, applicationID string) string {
	return filepath.Join(s.GetRoot(), tenantID, applicationID)
}

func (s *GitStorage) SaveApplication(application platform.HttpResponseApplication) error {
	mapped := storage.ConvertFromPlatformHttpResponseApplication(application)
	return s.SaveApplication2(mapped)
}

func (s *GitStorage) SaveApplication2(application storage.JSONApplication) error {
	applicationID := application.ID
	tenantID := application.TenantID
	logContext := s.logContext.WithFields(logrus.Fields{
		"method":        "SaveApplication",
		"customer":      tenantID,
		"applicationID": applicationID,
	})

	if err := s.Pull(); err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Pull")
		return err
	}

	filename, err := s.writeApplication(application)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("writeApplication")
		return err
	}

	err = s.CommitPathAndPush(filename, fmt.Sprintf("upsert application %s", applicationID))
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("CommitPathAndPush")
		return err
	}

	return nil
}

func (s *GitStorage) GetApplication(tenantID string, applicationID string) (storage.JSONApplication, error) {
	dir := s.GetApplicationDirectory(tenantID, applicationID)
	filename := filepath.Join(dir, "application.json")
	b, err := ioutil.ReadFile(filename)

	var application storage.JSONApplication
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return application, storage.ErrNotFound
		}
		return application, err
	}

	err = json.Unmarshal(b, &application)
	if err != nil {
		return application, err
	}
	return application, nil
}

func (s *GitStorage) GetApplications(customerID string) ([]storage.JSONApplication, error) {
	applicationIDs, err := s.discoverCustomerApplicationIds(customerID)
	applications := make([]storage.JSONApplication, 0)

	if err != nil {
		return applications, err
	}

	for _, applicationID := range applicationIDs {
		application, err := s.GetApplication(customerID, applicationID)
		if err != nil {
			s.logContext.WithFields(logrus.Fields{
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
			s.logContext.WithFields(logrus.Fields{
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

func (s *GitStorage) writeApplication(application storage.JSONApplication) (string, error) {
	customerID := application.TenantID
	applicationID := application.ID
	logContext := s.logContext.WithFields(logrus.Fields{
		"method":        "writeApplication",
		"customer":      application.TenantID,
		"applicationID": application.ID,
	})

	dir := s.GetApplicationDirectory(customerID, applicationID)
	filename := filepath.Join(dir, "application.json")
	err := s.writeToDisk(filename, application)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"filename": filename,
			"error":    err,
		}).Error("Failed to write 'application.json'")
		return filename, err
	}

	return filename, err
}
