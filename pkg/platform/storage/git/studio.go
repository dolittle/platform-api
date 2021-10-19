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
)

// SaveStudioConfigAndCommit pulls the remote, writes the studio.json file, commits the changes
// and pushes them to the remote
func (s *GitStorage) SaveStudioConfigAndCommit(tenantID string, config platform.StudioConfig) error {
	logContext := s.logContext.WithFields(logrus.Fields{
		"method":   "SaveStudioConfigAndCommit",
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

	filename, err := s.writeStudioConfig(tenantID, config)
	if err != nil {
		logContext.WithFields(log.Fields{
			"error": err,
		}).Error("writeStudioConfig")
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

	err = s.CommitAndPush(w, "upsert studio config")
	if err != nil {
		logContext.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to commit and push worktree")
		return err
	}

	return nil
}

// SaveStudioConfig pulls the remote repo and writes the new studio.json file without committing
// to the git repo
func (s *GitStorage) SaveStudioConfig(tenantID string, config platform.StudioConfig) error {
	logContext := s.logContext.WithFields(logrus.Fields{
		"method":   "SaveStudioConfig",
		"tenantID": tenantID,
	})

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

	_, err = s.writeStudioConfig(tenantID, config)
	if err != nil {
		logContext.WithFields(log.Fields{
			"error": err,
		}).Error("writeStudioConfig")
		return err
	}
	return nil
}

func (s *GitStorage) writeStudioConfig(tenantID string, config platform.StudioConfig) (string, error) {
	logContext := s.logContext.WithFields(log.Fields{
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
