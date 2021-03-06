// Code generated by mockery v2.12.2. DO NOT EDIT.

package k8s

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mock "github.com/stretchr/testify/mock"

	networkingv1 "k8s.io/api/networking/v1"

	testing "testing"

	v1 "k8s.io/api/rbac/v1"
)

// Repo is an autogenerated mock type for the Repo type
type Repo struct {
	mock.Mock
}

// AddSubjectToRoleBinding provides a mock function with given fields: namespace, name, subject
func (_m *Repo) AddSubjectToRoleBinding(namespace string, name string, subject v1.Subject) error {
	ret := _m.Called(namespace, name, subject)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, v1.Subject) error); ok {
		r0 = rf(namespace, name, subject)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetDeployment provides a mock function with given fields: namespace, environment, microserviceID
func (_m *Repo) GetDeployment(namespace string, environment string, microserviceID string) (appsv1.Deployment, error) {
	ret := _m.Called(namespace, environment, microserviceID)

	var r0 appsv1.Deployment
	if rf, ok := ret.Get(0).(func(string, string, string) appsv1.Deployment); ok {
		r0 = rf(namespace, environment, microserviceID)
	} else {
		r0 = ret.Get(0).(appsv1.Deployment)
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
func (_m *Repo) GetDeployments(namespace string) ([]appsv1.Deployment, error) {
	ret := _m.Called(namespace)

	var r0 []appsv1.Deployment
	if rf, ok := ret.Get(0).(func(string) []appsv1.Deployment); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]appsv1.Deployment)
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
func (_m *Repo) GetDeploymentsByEnvironmentWithMicroservice(namespace string, environment string) ([]appsv1.Deployment, error) {
	ret := _m.Called(namespace, environment)

	var r0 []appsv1.Deployment
	if rf, ok := ret.Get(0).(func(string, string) []appsv1.Deployment); ok {
		r0 = rf(namespace, environment)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]appsv1.Deployment)
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
func (_m *Repo) GetDeploymentsWithMicroservice(namespace string) ([]appsv1.Deployment, error) {
	ret := _m.Called(namespace)

	var r0 []appsv1.Deployment
	if rf, ok := ret.Get(0).(func(string) []appsv1.Deployment); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]appsv1.Deployment)
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
func (_m *Repo) GetDeploymentsWithOptions(namespace string, opts metav1.ListOptions) ([]appsv1.Deployment, error) {
	ret := _m.Called(namespace, opts)

	var r0 []appsv1.Deployment
	if rf, ok := ret.Get(0).(func(string, metav1.ListOptions) []appsv1.Deployment); ok {
		r0 = rf(namespace, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]appsv1.Deployment)
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

// GetIngresses provides a mock function with given fields: namespace
func (_m *Repo) GetIngresses(namespace string) ([]networkingv1.Ingress, error) {
	ret := _m.Called(namespace)

	var r0 []networkingv1.Ingress
	if rf, ok := ret.Get(0).(func(string) []networkingv1.Ingress); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]networkingv1.Ingress)
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
func (_m *Repo) GetIngressesByEnvironmentWithMicoservice(namespace string, environment string) ([]networkingv1.Ingress, error) {
	ret := _m.Called(namespace, environment)

	var r0 []networkingv1.Ingress
	if rf, ok := ret.Get(0).(func(string, string) []networkingv1.Ingress); ok {
		r0 = rf(namespace, environment)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]networkingv1.Ingress)
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
func (_m *Repo) GetIngressesWithOptions(namespace string, opts metav1.ListOptions) ([]networkingv1.Ingress, error) {
	ret := _m.Called(namespace, opts)

	var r0 []networkingv1.Ingress
	if rf, ok := ret.Get(0).(func(string, metav1.ListOptions) []networkingv1.Ingress); ok {
		r0 = rf(namespace, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]networkingv1.Ingress)
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

// GetNamespaces provides a mock function with given fields:
func (_m *Repo) GetNamespaces() ([]corev1.Namespace, error) {
	ret := _m.Called()

	var r0 []corev1.Namespace
	if rf, ok := ret.Get(0).(func() []corev1.Namespace); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]corev1.Namespace)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNamespacesWithApplication provides a mock function with given fields:
func (_m *Repo) GetNamespacesWithApplication() ([]corev1.Namespace, error) {
	ret := _m.Called()

	var r0 []corev1.Namespace
	if rf, ok := ret.Get(0).(func() []corev1.Namespace); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]corev1.Namespace)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNamespacesWithOptions provides a mock function with given fields: opts
func (_m *Repo) GetNamespacesWithOptions(opts metav1.ListOptions) ([]corev1.Namespace, error) {
	ret := _m.Called(opts)

	var r0 []corev1.Namespace
	if rf, ok := ret.Get(0).(func(metav1.ListOptions) []corev1.Namespace); ok {
		r0 = rf(opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]corev1.Namespace)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(metav1.ListOptions) error); ok {
		r1 = rf(opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetRoleBinding provides a mock function with given fields: namespace, name
func (_m *Repo) GetRoleBinding(namespace string, name string) (v1.RoleBinding, error) {
	ret := _m.Called(namespace, name)

	var r0 v1.RoleBinding
	if rf, ok := ret.Get(0).(func(string, string) v1.RoleBinding); ok {
		r0 = rf(namespace, name)
	} else {
		r0 = ret.Get(0).(v1.RoleBinding)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(namespace, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// HasUserAdminAccess provides a mock function with given fields: userID
func (_m *Repo) HasUserAdminAccess(userID string) (bool, error) {
	ret := _m.Called(userID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(userID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveSubjectToRoleBinding provides a mock function with given fields: namespace, name, subject
func (_m *Repo) RemoveSubjectToRoleBinding(namespace string, name string, subject v1.Subject) error {
	ret := _m.Called(namespace, name, subject)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, v1.Subject) error); ok {
		r0 = rf(namespace, name, subject)
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
