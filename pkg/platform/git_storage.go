package platform

import (
	"fmt"
	"log"
	"os"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type GitStorage struct {
	repo      *git.Repository
	directory string
	Repo      *git.Repository
	Directory string
}

func NewGitStorage(url string, directory string, privateKeyFile string) *GitStorage {
	s := &GitStorage{
		directory: directory,
		Directory: directory,
	}

	_, err := os.Stat(privateKeyFile)
	if err != nil {
		log.Fatal(err)
	}

	// Clone the given repository to the given directory
	publicKeys, err := ssh.NewPublicKeysFromFile("git", privateKeyFile, "")
	if err != nil {
		log.Fatalf("generate publickeys failed: %s\n", err.Error())
	}

	r, err := git.PlainClone(directory, false, &git.CloneOptions{
		// The intended use of a GitHub personal access token is in replace of your password
		// because access tokens can easily be revoked.
		// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
		Auth:     publicKeys,
		URL:      url,
		Progress: os.Stdout,
	})

	if err != nil {
		if err != git.ErrRepositoryAlreadyExists {
			log.Fatalf("cloning repo: %s\n", err.Error())
		}
		r, err = git.PlainOpen(directory)
		if err != nil {
			log.Fatalf("repo exists, opening: %s\n", err.Error())
		}
	}

	s.repo = r
	s.Repo = r
	return s
}

func (s *GitStorage) getApplicationDirectory(tenantID string, applicationID string) string {
	return fmt.Sprintf("%s/%s/%s", s.Directory, tenantID, applicationID)
}
