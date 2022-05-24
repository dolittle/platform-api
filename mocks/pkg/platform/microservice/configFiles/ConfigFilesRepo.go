// Code generated by mockery v2.12.2. DO NOT EDIT.

package configFiles

import (
	microserviceconfigFiles "github.com/dolittle/platform-api/pkg/platform/microservice/configFiles"
	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// ConfigFilesRepo is an autogenerated mock type for the ConfigFilesRepo type
type ConfigFilesRepo struct {
	mock.Mock
}

// AddEntryToConfigFiles provides a mock function with given fields: applicationID, environment, microserviceID, data
func (_m *ConfigFilesRepo) AddEntryToConfigFiles(applicationID string, environment string, microserviceID string, data microserviceconfigFiles.MicroserviceConfigFile) error {
	ret := _m.Called(applicationID, environment, microserviceID, data)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, microserviceconfigFiles.MicroserviceConfigFile) error); ok {
		r0 = rf(applicationID, environment, microserviceID, data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetConfigFilesNamesList provides a mock function with given fields: applicationID, environment, microserviceID
func (_m *ConfigFilesRepo) GetConfigFilesNamesList(applicationID string, environment string, microserviceID string) ([]string, error) {
	ret := _m.Called(applicationID, environment, microserviceID)

	var r0 []string
	if rf, ok := ret.Get(0).(func(string, string, string) []string); ok {
		r0 = rf(applicationID, environment, microserviceID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(applicationID, environment, microserviceID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveEntryFromConfigFiles provides a mock function with given fields: applicationID, environment, microserviceID, key
func (_m *ConfigFilesRepo) RemoveEntryFromConfigFiles(applicationID string, environment string, microserviceID string, key string) error {
	ret := _m.Called(applicationID, environment, microserviceID, key)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string) error); ok {
		r0 = rf(applicationID, environment, microserviceID, key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewConfigFilesRepo creates a new instance of ConfigFilesRepo. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewConfigFilesRepo(t testing.TB) *ConfigFilesRepo {
	mock := &ConfigFilesRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
