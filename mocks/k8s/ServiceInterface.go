// Code generated by mockery v2.16.0. DO NOT EDIT.

package mocks

import (
	context "context"
	k8s "ddm-admin-console/service/k8s"

	mock "github.com/stretchr/testify/mock"

	v1 "k8s.io/api/core/v1"
)

// ServiceInterface is an autogenerated mock type for the ServiceInterface type
type ServiceInterface struct {
	mock.Mock
}

// CanI provides a mock function with given fields: group, resource, verb, name
func (_m *ServiceInterface) CanI(group string, resource string, verb string, name string) (bool, error) {
	ret := _m.Called(group, resource, verb, name)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string, string, string) bool); ok {
		r0 = rf(group, resource, verb, name)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string) error); ok {
		r1 = rf(group, resource, verb, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetConfigMap provides a mock function with given fields: ctx, name, namespace
func (_m *ServiceInterface) GetConfigMap(ctx context.Context, name string, namespace string) (*v1.ConfigMap, error) {
	ret := _m.Called(ctx, name, namespace)

	var r0 *v1.ConfigMap
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *v1.ConfigMap); ok {
		r0 = rf(ctx, name, namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.ConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, name, namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSecret provides a mock function with given fields: name
func (_m *ServiceInterface) GetSecret(name string) (*v1.Secret, error) {
	ret := _m.Called(name)

	var r0 *v1.Secret
	if rf, ok := ret.Get(0).(func(string) *v1.Secret); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSecretFromNamespace provides a mock function with given fields: ctx, name, namespace
func (_m *ServiceInterface) GetSecretFromNamespace(ctx context.Context, name string, namespace string) (*v1.Secret, error) {
	ret := _m.Called(ctx, name, namespace)

	var r0 *v1.Secret
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *v1.Secret); ok {
		r0 = rf(ctx, name, namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, name, namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSecretKey provides a mock function with given fields: ctx, namespace, name, key
func (_m *ServiceInterface) GetSecretKey(ctx context.Context, namespace string, name string, key string) (string, error) {
	ret := _m.Called(ctx, namespace, name, key)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) string); ok {
		r0 = rf(ctx, namespace, name, key)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, namespace, name, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSecretKeys provides a mock function with given fields: ctx, namespace, name, keys
func (_m *ServiceInterface) GetSecretKeys(ctx context.Context, namespace string, name string, keys []string) (map[string]string, error) {
	ret := _m.Called(ctx, namespace, name, keys)

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func(context.Context, string, string, []string) map[string]string); ok {
		r0 = rf(ctx, namespace, name, keys)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, []string) error); ok {
		r1 = rf(ctx, namespace, name, keys)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RecreateSecret provides a mock function with given fields: secretName, data
func (_m *ServiceInterface) RecreateSecret(secretName string, data map[string][]byte) error {
	ret := _m.Called(secretName, data)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, map[string][]byte) error); ok {
		r0 = rf(secretName, data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ServiceForContext provides a mock function with given fields: ctx
func (_m *ServiceInterface) ServiceForContext(ctx context.Context) (k8s.ServiceInterface, error) {
	ret := _m.Called(ctx)

	var r0 k8s.ServiceInterface
	if rf, ok := ret.Get(0).(func(context.Context) k8s.ServiceInterface); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(k8s.ServiceInterface)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewServiceInterface interface {
	mock.TestingT
	Cleanup(func())
}

// NewServiceInterface creates a new instance of ServiceInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewServiceInterface(t mockConstructorTestingTNewServiceInterface) *ServiceInterface {
	mock := &ServiceInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
