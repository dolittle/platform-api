// Code generated by mockery v2.14.0. DO NOT EDIT.

package purchaseorderapi

import (
	k8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	microservicepurchaseorderapi "github.com/dolittle/platform-api/pkg/platform/microservice/purchaseorderapi"
	mock "github.com/stretchr/testify/mock"

	platform "github.com/dolittle/platform-api/pkg/platform"
)

// K8sResourceSpecFactory is an autogenerated mock type for the K8sResourceSpecFactory type
type K8sResourceSpecFactory struct {
	mock.Mock
}

// CreateAll provides a mock function with given fields: headImage, runtimeImage, k8sMicroservice, customerTenants, extra
func (_m *K8sResourceSpecFactory) CreateAll(headImage string, runtimeImage string, k8sMicroservice k8s.Microservice, customerTenants []platform.CustomerTenantInfo, extra platform.HttpInputPurchaseOrderExtra) microservicepurchaseorderapi.K8sResources {
	ret := _m.Called(headImage, runtimeImage, k8sMicroservice, customerTenants, extra)

	var r0 microservicepurchaseorderapi.K8sResources
	if rf, ok := ret.Get(0).(func(string, string, k8s.Microservice, []platform.CustomerTenantInfo, platform.HttpInputPurchaseOrderExtra) microservicepurchaseorderapi.K8sResources); ok {
		r0 = rf(headImage, runtimeImage, k8sMicroservice, customerTenants, extra)
	} else {
		r0 = ret.Get(0).(microservicepurchaseorderapi.K8sResources)
	}

	return r0
}

type mockConstructorTestingTNewK8sResourceSpecFactory interface {
	mock.TestingT
	Cleanup(func())
}

// NewK8sResourceSpecFactory creates a new instance of K8sResourceSpecFactory. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewK8sResourceSpecFactory(t mockConstructorTestingTNewK8sResourceSpecFactory) *K8sResourceSpecFactory {
	mock := &K8sResourceSpecFactory{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
