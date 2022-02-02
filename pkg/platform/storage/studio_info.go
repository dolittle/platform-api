package storage

import (
	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/sirupsen/logrus"
)

// TODO perhaps this should be a "server helper" and write responses
type StudioInfo struct {
	StudioConfig         platform.StudioConfig
	TerraformCustomer    platform.TerraformCustomer
	TerraformApplication platform.TerraformApplication
}

func GetStudioInfo(repo Repo, customerID string, applicationID string, logContext logrus.FieldLogger) (StudioInfo, error) {
	studioConfig, err := repo.GetStudioConfig(customerID)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"method":      "repo.GetStudioConfig",
			"error":       err,
			"customer_id": customerID,
		}).Error("Error while getting studio config")
		return StudioInfo{}, platform.ErrStudioInfoMissing
	}

	terraformCustomer, err := repo.GetTerraformTenant(customerID)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"method":      "repo.GetTerraformTenant",
			"error":       err,
			"customer_id": customerID,
		}).Error("Error while getting customer terraform")
		return StudioInfo{}, platform.ErrStudioInfoMissing
	}

	// Confirm the application has been created
	terraformApplication, err := repo.GetTerraformApplication(customerID, applicationID)
	if err != nil {
		// TODO handle not found
		logContext.WithFields(logrus.Fields{
			"method":         "repo.GetTerraformApplication",
			"error":          err,
			"customer_id":    customerID,
			"application_id": applicationID,
		}).Error("Error while getting application terraform")
		// This might be too aggressive and should be 404
		return StudioInfo{}, platform.ErrStudioInfoMissing
	}

	return StudioInfo{
		StudioConfig:         studioConfig,
		TerraformCustomer:    terraformCustomer,
		TerraformApplication: terraformApplication,
	}, nil
}
