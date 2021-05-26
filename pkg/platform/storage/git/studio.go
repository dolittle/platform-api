package git

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
)

func (s *GitStorage) GetStudioConfig(tenantID string) (platform.StudioConfig, error) {
	dir := s.GetTenantDirectory(tenantID)
	filename := fmt.Sprintf("%s/studio.json", dir)
	b, err := ioutil.ReadFile(filename)

	var config platform.StudioConfig
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(b, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}
