package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	gitSsh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
)

type GitStorageConfig struct {
	URL                 string
	Branch              string
	PrivateKey          string
	RepoRoot            string
	DirectoryOnly       bool
	DryRun              bool
	PlatformEnvironment string
}

type GitSync interface {
	Pull() error
}

type GitStorage struct {
	logContext logrus.FieldLogger
	Repo       *git.Repository
	Directory  string
	publicKeys *gitSsh.PublicKeys
	config     GitStorageConfig
}

func NewGitStorage(logContext logrus.FieldLogger, gitConfig GitStorageConfig) storage.Repo {
	directoryOnly := gitConfig.DirectoryOnly

	branch := plumbing.NewBranchReferenceName(gitConfig.Branch)

	// We remove the trailing path separator
	gitConfig.RepoRoot = strings.TrimSuffix(gitConfig.RepoRoot, string(os.PathSeparator))

	platformApiDir := filepath.Join(gitConfig.RepoRoot, "Source", "V3", "platform-api")

	s := &GitStorage{
		logContext: logContext.WithFields(logrus.Fields{
			"directoryOnly": directoryOnly,
			"gitRemote":     gitConfig.URL,
			"gitBranch":     gitConfig.Branch,
		}),
		Directory: platformApiDir,
		config:    gitConfig,
	}

	if directoryOnly {
		r, err := git.PlainOpenWithOptions(gitConfig.RepoRoot, &git.PlainOpenOptions{
			EnableDotGitCommonDir: true,
		})
		if err != nil {
			s.logContext.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("repo doesn't exist")
		}
		s.Repo = r
		return s
	}

	// Assume using remote repo
	_, err := os.Stat(gitConfig.PrivateKey)
	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("issue stat")
	}

	// Clone the given repository to the given directory
	s.publicKeys, err = gitSsh.NewPublicKeysFromFile("git", gitConfig.PrivateKey, "")
	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("generate publickeys failed")
	}

	// This is not ideal
	s.publicKeys.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	r, err := git.PlainClone(gitConfig.RepoRoot, false, &git.CloneOptions{
		// The intended use of a GitHub personal access token is in replace of your password
		// because access tokens can easily be revoked.
		// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
		Auth:          s.publicKeys,
		URL:           gitConfig.URL,
		Progress:      os.Stdout,
		ReferenceName: branch,
		// Neither of the below work
		//Depth:         1,
		// err object not found (doesnt work with either approach)
		//SingleBranch: true,
		// err empty git-upload-pack given
	})

	if err != nil {
		if err != git.ErrRepositoryAlreadyExists {
			s.logContext.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("cloning repo")
		}
		r, err = git.PlainOpenWithOptions(gitConfig.RepoRoot, &git.PlainOpenOptions{
			EnableDotGitCommonDir: true,
		})
		if err != nil {
			s.logContext.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("repo doesn't exist")
		}
	}

	s.Repo = r
	return s
}

func (s *GitStorage) GetDirectory() string {
	return s.Directory
}

// CommitPathAndPush adds the path to index, creates a commit, and pushes to the remote
func (s *GitStorage) CommitPathAndPush(path string, msg string) error {
	// I wonder if go-git has something built-in?
	path = strings.TrimPrefix(path, s.config.RepoRoot+string(os.PathSeparator))
	logContext := s.logContext.WithFields(logrus.Fields{
		"method": "CommitPathAndPush",
		"msg":    msg,
		"path":   path,
	})
	if s.config.DryRun {
		logContext.Info("dry-run configured, won't commit and push")
		return nil
	}

	w, err := s.Repo.Worktree()
	if err != nil {
		return err
	}

	err = w.AddWithOptions(&git.AddOptions{
		Path: path,
	})
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to add path to index")
		return err
	}

	commit, err := w.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Auto Platform",
			Email: "platform-auto@dolittle.com",
			When:  time.Now(),
		},
	})

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Commit")
		return err
	}

	// Prints the current HEAD to verify that all worked well.
	_, err = s.Repo.CommitObject(commit)

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("CommitObject")
		return err
	}

	// don't push if using a local repo
	// TODO this needs to be documented :P
	// Why do we have dryRun and DirectoryOnly? hmm
	if s.config.DirectoryOnly {
		return nil
	}

	err = s.Repo.Push(&git.PushOptions{
		Auth: s.publicKeys,
	})

	if err != nil {
		//if err == git.NoErrAlreadyUpToDate {}
		// If we have commited, this is a mistake
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Push")
		return err
	}

	logContext.Debug("Successfully pushed to remote")

	return err
}

// Pull pulls the latest from remote with the default Worktree.
// It returns nil on success
func (s *GitStorage) Pull() error {
	branchReference := plumbing.NewBranchReferenceName(s.config.Branch)
	logContext := s.logContext.WithFields(logrus.Fields{
		"method":          "Pull",
		"branchReference": branchReference,
	})
	if s.config.DirectoryOnly {
		logContext.Debug("Not pulling, repo is set to directoryOnly = true")
		return nil
	}

	worktree, err := s.Repo.Worktree()
	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Worktree")
		return err
	}

	err = worktree.Pull(&git.PullOptions{
		Auth:          s.publicKeys,
		ReferenceName: branchReference,
	})

	if err != nil && err != git.NoErrAlreadyUpToDate {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Pull")
		return err
	}

	return nil
}

func (s *GitStorage) IsAutomationEnabledWithStudioConfig(studioConfig platform.StudioConfig, applicationID string, environment string) bool {
	environment = strings.ToLower(environment)
	// If any of the entries == * disable all
	key := "*"
	if funk.ContainsString(studioConfig.DisabledEnvironments, key) {
		return false
	}

	key = fmt.Sprintf("%s/%s", applicationID, environment)
	return !funk.ContainsString(studioConfig.DisabledEnvironments, key)
}

func (s *GitStorage) GetRoot() string {
	return filepath.Join(s.Directory, s.config.PlatformEnvironment)
}
