package git

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func (s *GitStorage) GetTenantDirectory(tenantID string) string {
	return fmt.Sprintf("%s/%s", s.Directory, tenantID)
}

func (s *GitStorage) SaveTenant(tenant platform.TerraformCustomer) error {
	tenantID := tenant.GUID
	data, _ := json.Marshal(tenant)

	w, err := s.Repo.Worktree()
	if err != nil {
		return err
	}

	dir := s.GetTenantDirectory(tenantID)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Println("MkdirAll")
		return err
	}

	filename := fmt.Sprintf("%s/tenant.json", dir)
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		fmt.Println("writeFile")
		return err
	}

	// Adds the new file to the staging area.
	// Need to remove the prefix
	err = w.AddWithOptions(&git.AddOptions{
		Path: strings.TrimPrefix(filename, s.Directory+"/"),
	})

	if err != nil {
		fmt.Println("w.Add")
		return err
	}

	status, err := w.Status()
	if err != nil {
		fmt.Println("w.Status")
		return err
	}

	fmt.Println(status)

	commit, err := w.Commit("adding customer", &git.CommitOptions{
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
	_, err = s.Repo.CommitObject(commit)
	return err
}

func (s *GitStorage) GetTenant(tenantID string) (platform.TerraformCustomer, error) {
	dir := s.GetTenantDirectory(tenantID)
	filename := fmt.Sprintf("%s/tenant.json", dir)
	b, err := ioutil.ReadFile(filename)

	var tenant platform.TerraformCustomer
	if err != nil {
		return tenant, err
	}

	err = json.Unmarshal(b, &tenant)
	if err != nil {
		return tenant, err
	}
	return tenant, nil
}
