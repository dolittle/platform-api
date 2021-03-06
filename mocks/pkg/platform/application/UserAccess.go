// Code generated by mockery v2.12.2. DO NOT EDIT.

package application

import (
	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// UserAccess is an autogenerated mock type for the UserAccess type
type UserAccess struct {
	mock.Mock
}

// AddUser provides a mock function with given fields: customerID, applicationID, email
func (_m *UserAccess) AddUser(customerID string, applicationID string, email string) error {
	ret := _m.Called(customerID, applicationID, email)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string) error); ok {
		r0 = rf(customerID, applicationID, email)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetUsers provides a mock function with given fields: applicationID
func (_m *UserAccess) GetUsers(applicationID string) ([]string, error) {
	ret := _m.Called(applicationID)

	var r0 []string
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(applicationID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(applicationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveUser provides a mock function with given fields: applicationID, email
func (_m *UserAccess) RemoveUser(applicationID string, email string) error {
	ret := _m.Called(applicationID, email)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(applicationID, email)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewUserAccess creates a new instance of UserAccess. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewUserAccess(t testing.TB) *UserAccess {
	mock := &UserAccess{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
