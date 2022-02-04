package git

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/sirupsen/logrus"
)

// SaveStudioConfig pulls the remote, writes the studio.json file, commits the changes
// and pushes them to the remote
func (s *GitStorage) SaveStudioConfig(tenantID string, config platform.StudioConfig) error {
	logContext := s.logContext.WithFields(logrus.Fields{
		"method":   "SaveStudioConfig",
		"tenantID": tenantID,
	})

	if err := s.Pull(); err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Pull")
		return err
	}

	filename, err := s.writeStudioConfig(tenantID, config)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("writeStudioConfig")
		return err
	}

	err = s.CommitPathAndPush(filename, fmt.Sprintf("upsert studio config for customer %s", tenantID))
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("CommitPathAndPush")
		return err
	}

	return nil
}

func (s *GitStorage) writeStudioConfig(tenantID string, config platform.StudioConfig) (string, error) {
	logContext := s.logContext.WithFields(logrus.Fields{
		"method":   "writeStudioConfig",
		"customer": tenantID,
	})

	dir := s.GetTenantDirectory(tenantID)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}

	filename := filepath.Join(dir, "studio.json")
	data, _ := json.MarshalIndent(config, "", "  ")
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		logContext.WithFields(logrus.Fields{
			"error":    err,
			"filename": filename,
		}).Error("Failed to write to 'studio.json")
		return filename, err
	}
	return filename, nil
}

func (s *GitStorage) GetStudioConfig(tenantID string) (platform.StudioConfig, error) {
	dir := s.GetTenantDirectory(tenantID)
	filename := filepath.Join(dir, "studio.json")
	b, err := ioutil.ReadFile(filename)

	var config platform.StudioConfig
	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error":  err,
			"method": "GetStudioConfig",
		}).Error("no studio.json found")
		return config, err
	}

	err = json.Unmarshal(b, &config)
	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error":  err,
			"method": "GetStudioConfig",
		}).Error("couldn't parse studio.json")
		return config, err
	}

	return config, nil
}
