package git

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	git "github.com/go-git/go-git/v5"
)

func (s *GitStorage) SaveTerraformApplication(applicaiton platform.TerraformApplication) error {
	tenantID := applicaiton.Customer.GUID
	applicaitonID := applicaiton.GUID

	data, _ := json.Marshal(applicaiton)

	w, err := s.Repo.Worktree()
	if err != nil {
		return err
	}

	dir := s.GetApplicationDirectory(tenantID, applicaitonID)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Println("MkdirAll")
		return err
	}

	filename := fmt.Sprintf("%s/terraform.json", dir)
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

	_, err = w.Status()
	if err != nil {
		fmt.Println("w.Status")
		return err
	}
	err = s.CommitAndPush(w, "Adding customer")

	if err != nil {
		return err
	}

	return nil
}

func (s *GitStorage) GetTerraformApplication(tenantID string, applicationID string) (platform.TerraformApplication, error) {
	dir := s.GetApplicationDirectory(tenantID, applicationID)
	filename := fmt.Sprintf("%s/terraform.json", dir)
	b, err := ioutil.ReadFile(filename)

	var application platform.TerraformApplication
	if err != nil {
		return application, err
	}

	err = json.Unmarshal(b, &application)
	if err != nil {
		return application, err
	}
	return application, nil

}

func (s *GitStorage) SaveTerraformTenant(tenant platform.TerraformCustomer) error {
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

	_, err = w.Status()
	if err != nil {
		fmt.Println("w.Status")
		return err
	}
	err = s.CommitAndPush(w, "Adding customer")

	if err != nil {
		return err
	}

	return nil
}

func (s *GitStorage) GetTerraformTenant(tenantID string) (platform.TerraformCustomer, error) {
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