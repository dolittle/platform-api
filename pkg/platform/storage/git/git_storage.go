package git

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	gitSsh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
)

type GitStorageConfig struct {
	URL            string
	Branch         string
	PrivateKey     string
	LocalDirectory string
}
type GitStorage struct {
	logContext logrus.FieldLogger
	Repo       *git.Repository
	Directory  string
	publicKeys *gitSsh.PublicKeys
	config     GitStorageConfig
}

func NewGitStorage(logContext logrus.FieldLogger, gitConfig GitStorageConfig) *GitStorage {

	//func NewGitStorage(logContext logrus.FieldLogger, url string, directory string, branchName string, privateKeyFile string) *GitStorage {
	branch := plumbing.NewBranchReferenceName(gitConfig.Branch)

	s := &GitStorage{
		logContext: logContext,
		Directory:  gitConfig.LocalDirectory,
		config:     gitConfig,
	}

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

	r, err := git.PlainClone(gitConfig.LocalDirectory, false, &git.CloneOptions{
		// The intended use of a GitHub personal access token is in replace of your password
		// because access tokens can easily be revoked.
		// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
		Auth:          s.publicKeys,
		URL:           gitConfig.URL,
		Progress:      os.Stdout,
		ReferenceName: branch,
	})

	if err != nil {
		if err != git.ErrRepositoryAlreadyExists {
			s.logContext.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("cloning repo")
		}
		r, err = git.PlainOpen(gitConfig.LocalDirectory)
		if err != nil {
			s.logContext.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("repo exists")
		}
	}

	//w, err := r.Worktree()
	//if err != nil {
	//	log.Fatalf("repo exists, unable to get worktree: %s\n", err.Error())
	//}

	// Checkout
	//err = w.Checkout(&git.CheckoutOptions{
	//	Create: false,
	//	Branch: branch,
	//	Keep:   true,
	//})
	//
	//if err != nil {
	//	log.Fatalf("repo exists, unable to checkout branch: %s error %s\n", branchName, err.Error())
	//}

	s.Repo = r
	return s
}

func (s *GitStorage) CommitAndPush(w *git.Worktree, msg string) error {
	commit, err := w.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Auto Platform",
			Email: "platform-auto@dolittle.com",
			When:  time.Now(),
		},
	})

	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error": err,
			"msg":   msg,
		}).Error("Commit")
		return err
	}

	// Prints the current HEAD to verify that all worked well.
	_, err = s.Repo.CommitObject(commit)

	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error": err,
			"msg":   msg,
		}).Error("CommitObject")
		return err
	}

	err = s.Repo.Push(&git.PushOptions{
		Auth: s.publicKeys,
	})

	if err != nil {
		//if err == git.NoErrAlreadyUpToDate {}
		// If we have commited, this is a mistake
		s.logContext.WithFields(logrus.Fields{
			"error": err,
			"msg":   msg,
		}).Error("Push")
		return err
	}

	return err
}

func (s *GitStorage) Pull() error {
	// This code might need to be alot more fancy
	w, err := s.Repo.Worktree()
	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Worktree")
		return err
	}

	err = w.Pull(&git.PullOptions{
		Auth: s.publicKeys,
	})
	if err != nil {
		// Maybe trigger fatal?
		s.logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Pull")
		return err
	}

	return nil
}

func (s *GitStorage) IsAutomationEnabled(tenantID string, applicationID string, environment string) bool {
	environment = strings.ToLower(environment)
	studioConfig, err := s.GetStudioConfig(tenantID)
	if err != nil {
		// TODO maybe log this
		return false
	}

	if !studioConfig.AutomationEnabled {
		return false
	}

	key := fmt.Sprintf("%s/%s", applicationID, environment)
	return funk.ContainsString(studioConfig.AutomationEnvironments, key)
}

func (s *GitStorage) CheckAutomationEnabledViaCustomer(config platform.StudioConfig, applicationID string, environment string) bool {
	environment = strings.ToLower(environment)

	if !config.AutomationEnabled {
		return false
	}

	key := fmt.Sprintf("%s/%s", applicationID, environment)
	return funk.ContainsString(config.AutomationEnvironments, key)
}
