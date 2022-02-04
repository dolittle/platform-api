// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	platform "github.com/dolittle/platform-api/pkg/platform"
	storage "github.com/dolittle/platform-api/pkg/platform/storage"
	mock "github.com/stretchr/testify/mock"
)

// Repo is an autogenerated mock type for the Repo type
type Repo struct {
	mock.Mock
}

// DeleteBusinessMoment provides a mock function with given fields: tenantID, applicationID, environment, microserviceID, momentID
func (_m *Repo) DeleteBusinessMoment(tenantID string, applicationID string, environment string, microserviceID string, momentID string) error {
	ret := _m.Called(tenantID, applicationID, environment, microserviceID, momentID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string, string) error); ok {
		r0 = rf(tenantID, applicationID, environment, microserviceID, momentID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteBusinessMomentEntity provides a mock function with given fields: tenantID, applicationID, environment, microserviceID, entityID
func (_m *Repo) DeleteBusinessMomentEntity(tenantID string, applicationID string, environment string, microserviceID string, entityID string) error {
	ret := _m.Called(tenantID, applicationID, environment, microserviceID, entityID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string, string) error); ok {
		r0 = rf(tenantID, applicationID, environment, microserviceID, entityID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteMicroservice provides a mock function with given fields: tenantID, applicationID, environment, microserviceID
func (_m *Repo) DeleteMicroservice(tenantID string, applicationID string, environment string, microserviceID string) error {
	ret := _m.Called(tenantID, applicationID, environment, microserviceID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string) error); ok {
		r0 = rf(tenantID, applicationID, environment, microserviceID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetApplication provides a mock function with given fields: tenantID, applicationID
func (_m *Repo) GetApplication(tenantID string, applicationID string) (storage.JSONApplication2, error) {
	ret := _m.Called(tenantID, applicationID)

	var r0 storage.JSONApplication2
	if rf, ok := ret.Get(0).(func(string, string) storage.JSONApplication2); ok {
		r0 = rf(tenantID, applicationID)
	} else {
		r0 = ret.Get(0).(storage.JSONApplication2)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(tenantID, applicationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetApplications provides a mock function with given fields: tenantID
func (_m *Repo) GetApplications(tenantID string) ([]storage.JSONApplication2, error) {
	ret := _m.Called(tenantID)

	var r0 []storage.JSONApplication2
	if rf, ok := ret.Get(0).(func(string) []storage.JSONApplication2); ok {
		r0 = rf(tenantID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]storage.JSONApplication2)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(tenantID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBusinessMoments provides a mock function with given fields: tenantID, applicationID, environment
func (_m *Repo) GetBusinessMoments(tenantID string, applicationID string, environment string) (platform.HttpResponseBusinessMoments, error) {
	ret := _m.Called(tenantID, applicationID, environment)

	var r0 platform.HttpResponseBusinessMoments
	if rf, ok := ret.Get(0).(func(string, string, string) platform.HttpResponseBusinessMoments); ok {
		r0 = rf(tenantID, applicationID, environment)
	} else {
		r0 = ret.Get(0).(platform.HttpResponseBusinessMoments)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(tenantID, applicationID, environment)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetMicroservice provides a mock function with given fields: tenantID, applicationID, environment, microserviceID
func (_m *Repo) GetMicroservice(tenantID string, applicationID string, environment string, microserviceID string) ([]byte, error) {
	ret := _m.Called(tenantID, applicationID, environment, microserviceID)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string, string, string, string) []byte); ok {
		r0 = rf(tenantID, applicationID, environment, microserviceID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string) error); ok {
		r1 = rf(tenantID, applicationID, environment, microserviceID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetMicroservices provides a mock function with given fields: tenantID, applicationID
func (_m *Repo) GetMicroservices(tenantID string, applicationID string) ([]platform.HttpMicroserviceBase, error) {
	ret := _m.Called(tenantID, applicationID)

	var r0 []platform.HttpMicroserviceBase
	if rf, ok := ret.Get(0).(func(string, string) []platform.HttpMicroserviceBase); ok {
		r0 = rf(tenantID, applicationID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]platform.HttpMicroserviceBase)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(tenantID, applicationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetStudioConfig provides a mock function with given fields: tenantID
func (_m *Repo) GetStudioConfig(tenantID string) (platform.StudioConfig, error) {
	ret := _m.Called(tenantID)

	var r0 platform.StudioConfig
	if rf, ok := ret.Get(0).(func(string) platform.StudioConfig); ok {
		r0 = rf(tenantID)
	} else {
		r0 = ret.Get(0).(platform.StudioConfig)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(tenantID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTerraformApplication provides a mock function with given fields: tenantID, applicationID
func (_m *Repo) GetTerraformApplication(tenantID string, applicationID string) (platform.TerraformApplication, error) {
	ret := _m.Called(tenantID, applicationID)

	var r0 platform.TerraformApplication
	if rf, ok := ret.Get(0).(func(string, string) platform.TerraformApplication); ok {
		r0 = rf(tenantID, applicationID)
	} else {
		r0 = ret.Get(0).(platform.TerraformApplication)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(tenantID, applicationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTerraformTenant provides a mock function with given fields: tenantID
func (_m *Repo) GetTerraformTenant(tenantID string) (platform.TerraformCustomer, error) {
	ret := _m.Called(tenantID)

	var r0 platform.TerraformCustomer
	if rf, ok := ret.Get(0).(func(string) platform.TerraformCustomer); ok {
		r0 = rf(tenantID)
	} else {
		r0 = ret.Get(0).(platform.TerraformCustomer)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(tenantID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsAutomationEnabledWithStudioConfig provides a mock function with given fields: studioConfig, applicationID, environment
func (_m *Repo) IsAutomationEnabledWithStudioConfig(studioConfig platform.StudioConfig, applicationID string, environment string) bool {
	ret := _m.Called(studioConfig, applicationID, environment)

	var r0 bool
	if rf, ok := ret.Get(0).(func(platform.StudioConfig, string, string) bool); ok {
		r0 = rf(studioConfig, applicationID, environment)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// SaveApplication provides a mock function with given fields: application
func (_m *Repo) SaveApplication(application platform.HttpResponseApplication) error {
	ret := _m.Called(application)

	var r0 error
	if rf, ok := ret.Get(0).(func(platform.HttpResponseApplication) error); ok {
		r0 = rf(application)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveApplication2 provides a mock function with given fields: application
func (_m *Repo) SaveApplication2(application storage.JSONApplication2) error {
	ret := _m.Called(application)

	var r0 error
	if rf, ok := ret.Get(0).(func(storage.JSONApplication2) error); ok {
		r0 = rf(application)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveBusinessMoment provides a mock function with given fields: tenantID, input
func (_m *Repo) SaveBusinessMoment(tenantID string, input platform.HttpInputBusinessMoment) error {
	ret := _m.Called(tenantID, input)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, platform.HttpInputBusinessMoment) error); ok {
		r0 = rf(tenantID, input)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveBusinessMomentEntity provides a mock function with given fields: tenantID, input
func (_m *Repo) SaveBusinessMomentEntity(tenantID string, input platform.HttpInputBusinessMomentEntity) error {
	ret := _m.Called(tenantID, input)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, platform.HttpInputBusinessMomentEntity) error); ok {
		r0 = rf(tenantID, input)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveMicroservice provides a mock function with given fields: tenantID, applicationID, environment, microserviceID, data
func (_m *Repo) SaveMicroservice(tenantID string, applicationID string, environment string, microserviceID string, data interface{}) error {
	ret := _m.Called(tenantID, applicationID, environment, microserviceID, data)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string, interface{}) error); ok {
		r0 = rf(tenantID, applicationID, environment, microserviceID, data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveStudioConfig provides a mock function with given fields: tenantID, config
func (_m *Repo) SaveStudioConfig(tenantID string, config platform.StudioConfig) error {
	ret := _m.Called(tenantID, config)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, platform.StudioConfig) error); ok {
		r0 = rf(tenantID, config)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveTerraformApplication provides a mock function with given fields: application
func (_m *Repo) SaveTerraformApplication(application platform.TerraformApplication) error {
	ret := _m.Called(application)

	var r0 error
	if rf, ok := ret.Get(0).(func(platform.TerraformApplication) error); ok {
		r0 = rf(application)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveTerraformTenant provides a mock function with given fields: tenant
func (_m *Repo) SaveTerraformTenant(tenant platform.TerraformCustomer) error {
	ret := _m.Called(tenant)

	var r0 error
	if rf, ok := ret.Get(0).(func(platform.TerraformCustomer) error); ok {
		r0 = rf(tenant)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
