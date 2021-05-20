package git

import (
	"log"
	"os"
	"time"

	"golang.org/x/crypto/ssh"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	gitSsh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type GitStorage struct {
	repo      *git.Repository
	directory string
	Repo      *git.Repository
	Directory string
}

func NewGitStorage(url string, directory string, branchName string, privateKeyFile string) *GitStorage {
	branch := plumbing.NewBranchReferenceName(branchName)

	s := &GitStorage{
		directory: directory,
		Directory: directory,
	}

	_, err := os.Stat(privateKeyFile)
	if err != nil {
		log.Fatal(err)
	}

	// Clone the given repository to the given directory
	publicKeys, err := gitSsh.NewPublicKeysFromFile("git", privateKeyFile, "")
	if err != nil {
		log.Fatalf("generate publickeys failed: %s\n", err.Error())
	}

	// This is not ideal
	publicKeys.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	r, err := git.PlainClone(directory, false, &git.CloneOptions{
		// The intended use of a GitHub personal access token is in replace of your password
		// because access tokens can easily be revoked.
		// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
		Auth:          publicKeys,
		URL:           url,
		Progress:      os.Stdout,
		ReferenceName: branch,
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

	s.repo = r
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
		return err
	}

	// Prints the current HEAD to verify that all worked well.
	_, err = s.Repo.CommitObject(commit)

	if err != nil {
		return err
	}

	err = s.Repo.Push(&git.PushOptions{})
	if err != nil {
		//if err == git.NoErrAlreadyUpToDate {}
		// If we have commited, this is a mistake
		return err
	}

	return err
}

func (s *GitStorage) Pull() error {
	// This code might need to be alot more fancy
	w, err := s.Repo.Worktree()
	if err != nil {
		// Maybe trigger fatal?
		return err
	}

	err = w.Pull(&git.PullOptions{})
	if err != nil {
		// Maybe trigger fatal?
		return err
	}

	return nil
}
