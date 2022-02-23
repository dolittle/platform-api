// Code generated by mockery v2.9.4. DO NOT EDIT.

package k8s

import (
	mock "github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/apps/v1"
)

// RepoDeployment is an autogenerated mock type for the RepoDeployment type
type RepoDeployment struct {
	mock.Mock
}

// GetDeployment provides a mock function with given fields: namespace, environment, microserviceID
func (_m *RepoDeployment) GetDeployment(namespace string, environment string, microserviceID string) (v1.Deployment, error) {
	ret := _m.Called(namespace, environment, microserviceID)

	var r0 v1.Deployment
	if rf, ok := ret.Get(0).(func(string, string, string) v1.Deployment); ok {
		r0 = rf(namespace, environment, microserviceID)
	} else {
		r0 = ret.Get(0).(v1.Deployment)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(namespace, environment, microserviceID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDeployments provides a mock function with given fields: namespace
func (_m *RepoDeployment) GetDeployments(namespace string) ([]v1.Deployment, error) {
	ret := _m.Called(namespace)

	var r0 []v1.Deployment
	if rf, ok := ret.Get(0).(func(string) []v1.Deployment); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]v1.Deployment)
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

// GetDeploymentsByEnvironmentWithMicroservice provides a mock function with given fields: namespace, environment
func (_m *RepoDeployment) GetDeploymentsByEnvironmentWithMicroservice(namespace string, environment string) ([]v1.Deployment, error) {
	ret := _m.Called(namespace, environment)

	var r0 []v1.Deployment
	if rf, ok := ret.Get(0).(func(string, string) []v1.Deployment); ok {
		r0 = rf(namespace, environment)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]v1.Deployment)
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

// GetDeploymentsWithMicroservice provides a mock function with given fields: namespace
func (_m *RepoDeployment) GetDeploymentsWithMicroservice(namespace string) ([]v1.Deployment, error) {
	ret := _m.Called(namespace)

	var r0 []v1.Deployment
	if rf, ok := ret.Get(0).(func(string) []v1.Deployment); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]v1.Deployment)
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

// GetDeploymentsWithOptions provides a mock function with given fields: namespace, opts
func (_m *RepoDeployment) GetDeploymentsWithOptions(namespace string, opts metav1.ListOptions) ([]v1.Deployment, error) {
	ret := _m.Called(namespace, opts)

	var r0 []v1.Deployment
	if rf, ok := ret.Get(0).(func(string, metav1.ListOptions) []v1.Deployment); ok {
		r0 = rf(namespace, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]v1.Deployment)
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
