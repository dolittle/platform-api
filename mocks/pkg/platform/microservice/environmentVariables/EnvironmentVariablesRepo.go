// Code generated by mockery v2.12.2. DO NOT EDIT.

package environmentVariables

import (
	testing "testing"

	platform "github.com/dolittle/platform-api/pkg/platform"
	mock "github.com/stretchr/testify/mock"
)

// EnvironmentVariablesRepo is an autogenerated mock type for the EnvironmentVariablesRepo type
type EnvironmentVariablesRepo struct {
	mock.Mock
}

// GetEnvironmentVariables provides a mock function with given fields: applicationID, environment, microserviceID
func (_m *EnvironmentVariablesRepo) GetEnvironmentVariables(applicationID string, environment string, microserviceID string) ([]platform.StudioEnvironmentVariable, error) {
	ret := _m.Called(applicationID, environment, microserviceID)

	var r0 []platform.StudioEnvironmentVariable
	if rf, ok := ret.Get(0).(func(string, string, string) []platform.StudioEnvironmentVariable); ok {
		r0 = rf(applicationID, environment, microserviceID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]platform.StudioEnvironmentVariable)
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

// UpdateEnvironmentVariables provides a mock function with given fields: applicationID, environment, microserviceID, data
func (_m *EnvironmentVariablesRepo) UpdateEnvironmentVariables(applicationID string, environment string, microserviceID string, data []platform.StudioEnvironmentVariable) error {
	ret := _m.Called(applicationID, environment, microserviceID, data)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, []platform.StudioEnvironmentVariable) error); ok {
		r0 = rf(applicationID, environment, microserviceID, data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewEnvironmentVariablesRepo creates a new instance of EnvironmentVariablesRepo. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewEnvironmentVariablesRepo(t testing.TB) *EnvironmentVariablesRepo {
	mock := &EnvironmentVariablesRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
