// Code generated by mockery v2.12.2. DO NOT EDIT.

package rawdatalog

import (
	testing "testing"

	pkgrawdatalog "github.com/dolittle/platform-api/pkg/rawdatalog"
	mock "github.com/stretchr/testify/mock"
)

// Repo is an autogenerated mock type for the Repo type
type Repo struct {
	mock.Mock
}

// Write provides a mock function with given fields: topic, moment
func (_m *Repo) Write(topic string, moment pkgrawdatalog.RawMoment) error {
	ret := _m.Called(topic, moment)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, pkgrawdatalog.RawMoment) error); ok {
		r0 = rf(topic, moment)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewRepo creates a new instance of Repo. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewRepo(t testing.TB) *Repo {
	mock := &Repo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
