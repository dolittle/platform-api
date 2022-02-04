package git

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/sirupsen/logrus"
)

func (s *GitStorage) GetTenantDirectory(tenantID string) string {
	return filepath.Join(s.GetRoot(), tenantID)
}

func (s *GitStorage) GetCustomers() ([]platform.Customer, error) {
	pathToCustomers := []string{}

	dir := filepath.Join(s.Directory, s.config.PlatformEnvironment)

	// TODO change to fs when gone to 1.16
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			_path := strings.TrimPrefix(path, dir)

			parts := strings.Split(_path, string(os.PathSeparator))
			if len(parts) > 2 {
				return filepath.SkipDir
			}
			return nil
		}

		if info.Name() != "customer.json" {
			return nil
		}

		pathToCustomers = append(pathToCustomers, path)
		return nil
	})

	customers := make([]platform.Customer, 0)
	for _, pathToCustomer := range pathToCustomers {
		b, err := ioutil.ReadFile(pathToCustomer)

		if err != nil {
			continue
		}
		var customer platform.Customer
		_ = json.Unmarshal(b, &customer)
		customers = append(customers, customer)
	}
	return customers, nil
}

func (s *GitStorage) SaveCustomer(customer storage.JSONCustomer) error {
	customerID := customer.ID

	logContext := s.logContext.WithFields(logrus.Fields{
		"method":     "SaveCustomer",
		"customerID": customerID,
	})

	if err := s.Pull(); err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Pull")
		return err
	}

	dir := s.GetTenantDirectory(customerID)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	filename := filepath.Join(dir, "customer.json")
	err = s.writeToDisk(filename, customer)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("write to disk")
		return err
	}

	err = s.CommitPathAndPush(filename, fmt.Sprintf("upsert customer %s", customerID))
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("CommitPathAndPush")
		return err
	}

	return nil
}

func (s *GitStorage) writeToDisk(filename string, data interface{}) error {
	dir := path.Dir(filename)
	b, _ := json.MarshalIndent(data, "", " ")

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, b, 0644)
}
