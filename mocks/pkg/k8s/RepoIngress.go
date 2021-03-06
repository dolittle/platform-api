// Code generated by mockery v2.12.2. DO NOT EDIT.

package k8s

import (
	mock "github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	testing "testing"

	v1 "k8s.io/api/networking/v1"
)

// RepoIngress is an autogenerated mock type for the RepoIngress type
type RepoIngress struct {
	mock.Mock
}

// GetIngresses provides a mock function with given fields: namespace
func (_m *RepoIngress) GetIngresses(namespace string) ([]v1.Ingress, error) {
	ret := _m.Called(namespace)

	var r0 []v1.Ingress
	if rf, ok := ret.Get(0).(func(string) []v1.Ingress); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]v1.Ingress)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetIngressesByEnvironmentWithMicoservice provides a mock function with given fields: namespace, environment
func (_m *RepoIngress) GetIngressesByEnvironmentWithMicoservice(namespace string, environment string) ([]v1.Ingress, error) {
	ret := _m.Called(namespace, environment)

	var r0 []v1.Ingress
	if rf, ok := ret.Get(0).(func(string, string) []v1.Ingress); ok {
		r0 = rf(namespace, environment)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]v1.Ingress)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(namespace, environment)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetIngressesWithOptions provides a mock function with given fields: namespace, opts
func (_m *RepoIngress) GetIngressesWithOptions(namespace string, opts metav1.ListOptions) ([]v1.Ingress, error) {
	ret := _m.Called(namespace, opts)

	var r0 []v1.Ingress
	if rf, ok := ret.Get(0).(func(string, metav1.ListOptions) []v1.Ingress); ok {
		r0 = rf(namespace, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]v1.Ingress)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, metav1.ListOptions) error); ok {
		r1 = rf(namespace, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewRepoIngress creates a new instance of RepoIngress. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewRepoIngress(t testing.TB) *RepoIngress {
	mock := &RepoIngress{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
