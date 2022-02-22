// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	platform "github.com/dolittle/platform-api/pkg/platform"
	mock "github.com/stretchr/testify/mock"
)

// RepoMicroservice is an autogenerated mock type for the RepoMicroservice type
type RepoMicroservice struct {
	mock.Mock
}

// DeleteMicroservice provides a mock function with given fields: customerID, applicationID, environment, microserviceID
func (_m *RepoMicroservice) DeleteMicroservice(customerID string, applicationID string, environment string, microserviceID string) error {
	ret := _m.Called(customerID, applicationID, environment, microserviceID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string) error); ok {
		r0 = rf(customerID, applicationID, environment, microserviceID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetMicroservice provides a mock function with given fields: customerID, applicationID, environment, microserviceID
func (_m *RepoMicroservice) GetMicroservice(customerID string, applicationID string, environment string, microserviceID string) ([]byte, error) {
	ret := _m.Called(customerID, applicationID, environment, microserviceID)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string, string, string, string) []byte); ok {
		r0 = rf(customerID, applicationID, environment, microserviceID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string) error); ok {
		r1 = rf(customerID, applicationID, environment, microserviceID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetMicroservices provides a mock function with given fields: customerID, applicationID
func (_m *RepoMicroservice) GetMicroservices(customerID string, applicationID string) ([]platform.HttpMicroserviceBase, error) {
	ret := _m.Called(customerID, applicationID)

	var r0 []platform.HttpMicroserviceBase
	if rf, ok := ret.Get(0).(func(string, string) []platform.HttpMicroserviceBase); ok {
		r0 = rf(customerID, applicationID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]platform.HttpMicroserviceBase)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(customerID, applicationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SaveMicroservice provides a mock function with given fields: customerID, applicationID, environment, microserviceID, data
func (_m *RepoMicroservice) SaveMicroservice(customerID string, applicationID string, environment string, microserviceID string, data interface{}) error {
	ret := _m.Called(customerID, applicationID, environment, microserviceID, data)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string, interface{}) error); ok {
		r0 = rf(customerID, applicationID, environment, microserviceID, data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
