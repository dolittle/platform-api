// Code generated by mockery v2.12.2. DO NOT EDIT.

package user

import (
	testing "testing"

	platformuser "github.com/dolittle/platform-api/pkg/platform/user"
	mock "github.com/stretchr/testify/mock"
)

// UserActiveDirectory is an autogenerated mock type for the UserActiveDirectory type
type UserActiveDirectory struct {
	mock.Mock
}

// AddUserToGroup provides a mock function with given fields: userID, groupID
func (_m *UserActiveDirectory) AddUserToGroup(userID string, groupID string) error {
	ret := _m.Called(userID, groupID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(userID, groupID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AddUserToGroupByEmail provides a mock function with given fields: email, groupID
func (_m *UserActiveDirectory) AddUserToGroupByEmail(email string, groupID string) error {
	ret := _m.Called(email, groupID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(email, groupID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetGroupIDByApplicationID provides a mock function with given fields: applicationID
func (_m *UserActiveDirectory) GetGroupIDByApplicationID(applicationID string) (string, error) {
	ret := _m.Called(applicationID)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(applicationID)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(applicationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUserIDByEmail provides a mock function with given fields: email
func (_m *UserActiveDirectory) GetUserIDByEmail(email string) (string, error) {
	ret := _m.Called(email)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(email)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(email)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUsersInApplication provides a mock function with given fields: groupID
func (_m *UserActiveDirectory) GetUsersInApplication(groupID string) ([]platformuser.ActiveDirectoryUserInfo, error) {
	ret := _m.Called(groupID)

	var r0 []platformuser.ActiveDirectoryUserInfo
	if rf, ok := ret.Get(0).(func(string) []platformuser.ActiveDirectoryUserInfo); ok {
		r0 = rf(groupID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]platformuser.ActiveDirectoryUserInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(groupID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveUserFromGroup provides a mock function with given fields: userID, groupID
func (_m *UserActiveDirectory) RemoveUserFromGroup(userID string, groupID string) error {
	ret := _m.Called(userID, groupID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(userID, groupID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveUserFromGroupByEmail provides a mock function with given fields: email, groupID
func (_m *UserActiveDirectory) RemoveUserFromGroupByEmail(email string, groupID string) error {
	ret := _m.Called(email, groupID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(email, groupID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewUserActiveDirectory creates a new instance of UserActiveDirectory. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewUserActiveDirectory(t testing.TB) *UserActiveDirectory {
	mock := &UserActiveDirectory{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
