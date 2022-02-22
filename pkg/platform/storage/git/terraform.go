package git

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/sirupsen/logrus"
)

func (s *GitStorage) SaveTerraformApplication(application platform.TerraformApplication) error {
	tenantID := application.Customer.GUID
	applicationID := application.GUID

	data, _ := json.MarshalIndent(application, "", "  ")

	if err := s.Pull(); err != nil {
		s.logContext.WithFields(logrus.Fields{
			"method": "SaveTerraformApplication",
			"error":  err,
		}).Error("Pull")
		return err
	}

	dir := s.GetApplicationDirectory(tenantID, applicationID)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Println("MkdirAll")
		return err
	}

	filename := filepath.Join(dir, "terraform.json")
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		fmt.Println("writeFile")
		return err
	}

	err = s.CommitPathAndPush(filename, fmt.Sprintf("Adding application %s for customer %s", applicationID, tenantID))

	if err != nil {
		return err
	}

	return nil
}

func (s *GitStorage) GetTerraformApplication(customerID string, applicationID string) (platform.TerraformApplication, error) {
	dir := s.GetApplicationDirectory(customerID, applicationID)
	filename := filepath.Join(dir, "terraform.json")
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
	data, _ := json.MarshalIndent(tenant, "", "  ")

	if err := s.Pull(); err != nil {
		s.logContext.WithFields(logrus.Fields{
			"method": "SaveTerraformTenant",
			"error":  err,
		}).Error("Pull")
		return err
	}

	dir := s.GetCustomerDirectory(tenantID)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Println("MkdirAll")
		return err
	}

	filename := filepath.Join(dir, "tenant.json")
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		fmt.Println("writeFile")
		return err
	}

	err = s.CommitPathAndPush(filename, fmt.Sprintf("Adding customer %s", tenantID))

	if err != nil {
		return err
	}

	return nil
}

func (s *GitStorage) GetTerraformTenant(customerID string) (platform.TerraformCustomer, error) {
	dir := s.GetCustomerDirectory(customerID)
	filename := filepath.Join(dir, "tenant.json")
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
