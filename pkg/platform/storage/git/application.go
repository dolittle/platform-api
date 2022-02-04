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
	"github.com/thoas/go-funk"
)

func (s *GitStorage) GetApplicationDirectory(tenantID string, applicationID string) string {
	return filepath.Join(s.GetRoot(), tenantID, applicationID)
}

func (s *GitStorage) GetApplications(customerID string) ([]platform.HttpResponseApplication, error) {
	stored, err := s.GetApplications2(customerID)

	if err != nil {
		return make([]platform.HttpResponseApplication, 0), err
	}

	applications := funk.Map(stored, func(application storage.JSONApplication2) platform.HttpResponseApplication {
		return storage.ConvertFromJSONApplication2(application)
	}).([]platform.HttpResponseApplication)

	return applications, nil
}

func (s *GitStorage) SaveApplication(application platform.HttpResponseApplication) error {
	mapped := storage.ConvertFromPlatformHttpResponseApplication(application)
	return s.SaveApplication2(mapped)
}

func (s *GitStorage) SaveApplication2(application storage.JSONApplication2) error {
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

func (s *GitStorage) GetApplication2(tenantID string, applicationID string) (storage.JSONApplication2, error) {
	dir := s.GetApplicationDirectory(tenantID, applicationID)
	filename := filepath.Join(dir, "application.json")
	b, err := ioutil.ReadFile(filename)

	var application storage.JSONApplication2
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

func (s *GitStorage) GetApplications2(customerID string) ([]storage.JSONApplication2, error) {
	applicationIDs, err := s.discoverCustomerApplicationIds(customerID)
	applications := make([]storage.JSONApplication2, 0)

	if err != nil {
		return applications, err
	}

	for _, applicationID := range applicationIDs {
		application, err := s.GetApplication2(customerID, applicationID)
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

func (s *GitStorage) writeApplication(application storage.JSONApplication2) (string, error) {
	customerID := application.TenantID
	applicationID := application.ID
	logContext := s.logContext.WithFields(logrus.Fields{
		"method":        "writeApplication",
		"customer":      application.TenantID,
		"applicationID": application.ID,
	})

	data, _ := json.MarshalIndent(application, "", " ")

	dir := s.GetApplicationDirectory(customerID, applicationID)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}

	filename := filepath.Join(dir, "application.json")
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to write 'application.json'")
		return filename, err
	}

	return filename, err
}
