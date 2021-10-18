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
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func (s *GitStorage) SaveStudioConfig(tenantID string, config platform.StudioConfig) error {
	logContext := s.logContext.WithFields(logrus.Fields{
		"method":   "SaveStudioConfig",
		"tenantID": tenantID,
	})

	w, err := s.Repo.Worktree()
	if err != nil {
		return err
	}
	if err := s.PullWithWorktree(w); err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Pull")
		return err
	}
	dir := s.GetTenantDirectory(tenantID)
	filename := filepath.Join(dir, "studio.json")
	data, _ := json.MarshalIndent(config, "", "  ")
	if err = ioutil.WriteFile(filename, data, 0644); err != nil {
		logContext.WithFields(logrus.Fields{
			"error":    err,
			"filename": filename,
		}).Error("Failed to write to 'studio.json")
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

	err = s.CommitAndPush(w, "upsert studio.json")
	if err != nil {
		logContext.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to commit and push worktree")
		return err
	}

	return nil
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

// CreateDefaultStudioConfig creates a studio.json file with default values
// set to enable automation and overwriting for that customer.
// The given applications will have all of their environments enabled for automation too.
func (s *GitStorage) CreateDefaultStudioConfig(customerID string, applications []platform.HttpResponseApplication) error {
	var environments []string

	for _, application := range applications {
		for _, environment := range application.Environments {
			applicationWithEnvironment := fmt.Sprintf("%s/%s", application.ID, strings.ToLower(environment.Name))
			environments = append(environments, applicationWithEnvironment)
		}
	}

	studioConfig := platform.StudioConfig{
		BuildOverwrite:         true,
		AutomationEnabled:      true,
		AutomationEnvironments: environments,
	}

	return s.SaveStudioConfig(customerID, studioConfig)
}
