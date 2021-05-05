package application

import (
	"encoding/json"
	"fmt"
	"strings"

	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type gitStorage struct {
	repo      *git.Repository
	directory string
}

func NewGitStorage(url string, directory string, privateKeyFile string) *gitStorage {
	s := &gitStorage{
		directory: directory,
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
	fmt.Println(s.repo)
	return s
}

func (s *gitStorage) Write(tenantID string, applicationID string, data []byte) error {
	fmt.Printf("Write %s.json to file", applicationID)

	w, err := s.repo.Worktree()
	if err != nil {
		return err
	}

	// TODO actually build structure
	suffix := fmt.Sprintf("application_%s_%s.json", tenantID, applicationID)

	filename := filepath.Join(s.directory, suffix)
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}

	// Adds the new file to the staging area.
	_, err = w.Add(suffix)
	if err != nil {
		return err
	}

	status, err := w.Status()
	if err != nil {
		return err
	}

	fmt.Println(status)

	commit, err := w.Commit("example go-git commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  time.Now(),
		},
	})

	if err != nil {
		return err
	}

	// Prints the current HEAD to verify that all worked well.
	_, err = s.repo.CommitObject(commit)
	return err
}

func (s *gitStorage) Read(tenantID string, applicationID string) ([]byte, error) {
	suffix := fmt.Sprintf("application_%s_%s.json", tenantID, applicationID)
	filename := filepath.Join(s.directory, suffix)
	return ioutil.ReadFile(filename)
}

func (s *gitStorage) GetAll(tenantID string) ([]Application, error) {
	files := []string{}

	// TODO change
	rootDirectory := s.directory + "/"
	// TODO change to fs when gone to 1.16
	err := filepath.Walk(rootDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		// TODO come back to this
		//if info.IsDir() {
		//	fmt.Printf("skipping a dir without errors: %+v \n", info.Name())
		//	return filepath.SkipDir
		//}

		if !strings.HasSuffix(info.Name(), ".json") {
			return nil
		}

		if !strings.HasPrefix(info.Name(), "application_") {
			return nil
		}

		files = append(files, path)
		return nil
	})

	applications := make([]Application, 0)

	if err != nil {
		return applications, err
	}

	for _, filename := range files {
		var application Application
		b, _ := ioutil.ReadFile(filename)
		json.Unmarshal(b, &application)
		applications = append(applications, application)
	}

	return applications, nil
}
