package git

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/sirupsen/logrus"
)

func (s *GitStorage) SaveStudioConfig(tenantID string, config platform.StudioConfig) error {
	if err := s.Pull(); err != nil {
		s.logContext.WithFields(logrus.Fields{
			"method": "SaveStudioConfig",
			"error":  err,
		}).Error("Pull")
		return err
	}
	dir := s.GetTenantDirectory(tenantID)
	filename := filepath.Join(dir, "studio.json")
	data, _ := json.MarshalIndent(config, "", "  ")
	return ioutil.WriteFile(filename, data, 0644)
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
		}).Error("lookup getting studio.json")
		config.BuildOverwrite = true
		config.AutomationEnabled = false
		config.AutomationEnvironments = make([]string, 0)
		return config, nil
	}

	err = json.Unmarshal(b, &config)
	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error":  err,
			"method": "GetStudioConfig",
		}).Error("parsing json")
		config.BuildOverwrite = false
		config.AutomationEnabled = false
		config.AutomationEnvironments = make([]string, 0)
		return config, nil
	}

	return config, nil
}
