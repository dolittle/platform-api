package git

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/sirupsen/logrus"
)

func (s *GitStorage) SaveStudioConfig(tenantID string, config platform.StudioConfig) error {
	dir := s.GetTenantDirectory(tenantID)
	filename := fmt.Sprintf("%s/studio.json", dir)
	data, _ := json.Marshal(config)
	return ioutil.WriteFile(filename, data, 0644)
}

func (s *GitStorage) GetStudioConfig(tenantID string) (platform.StudioConfig, error) {
	dir := s.GetTenantDirectory(tenantID)
	filename := fmt.Sprintf("%s/studio.json", dir)
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
