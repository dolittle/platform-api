// Code generated by mockery v2.12.2. DO NOT EDIT.

package parser

import (
	k8s "github.com/dolittle/platform-api/pkg/platform/microservice/k8s"
	errors "k8s.io/apimachinery/pkg/api/errors"

	mock "github.com/stretchr/testify/mock"

	platform "github.com/dolittle/platform-api/pkg/platform"

	testing "testing"
)

// Parser is an autogenerated mock type for the Parser type
type Parser struct {
	mock.Mock
}

// Parse provides a mock function with given fields: requestBytes, microservice, applicationInfo
func (_m *Parser) Parse(requestBytes []byte, microservice platform.Microservice, applicationInfo platform.Application) (k8s.MicroserviceK8sInfo, *errors.StatusError) {
	ret := _m.Called(requestBytes, microservice, applicationInfo)

	var r0 k8s.MicroserviceK8sInfo
	if rf, ok := ret.Get(0).(func([]byte, platform.Microservice, platform.Application) k8s.MicroserviceK8sInfo); ok {
		r0 = rf(requestBytes, microservice, applicationInfo)
	} else {
		r0 = ret.Get(0).(k8s.MicroserviceK8sInfo)
	}

	var r1 *errors.StatusError
	if rf, ok := ret.Get(1).(func([]byte, platform.Microservice, platform.Application) *errors.StatusError); ok {
		r1 = rf(requestBytes, microservice, applicationInfo)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*errors.StatusError)
		}
	}

	return r0, r1
}

// NewParser creates a new instance of Parser. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewParser(t testing.TB) *Parser {
	mock := &Parser{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
