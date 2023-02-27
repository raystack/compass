// Code generated by mockery v2.20.2. DO NOT EDIT.

package mocks

import (
	context "context"

	namespace "github.com/odpf/compass/core/namespace"
	mock "github.com/stretchr/testify/mock"
)

// NamespaceDiscoveryRepository is an autogenerated mock type for the DiscoveryRepository type
type NamespaceDiscoveryRepository struct {
	mock.Mock
}

type NamespaceDiscoveryRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *NamespaceDiscoveryRepository) EXPECT() *NamespaceDiscoveryRepository_Expecter {
	return &NamespaceDiscoveryRepository_Expecter{mock: &_m.Mock}
}

// CreateNamespace provides a mock function with given fields: _a0, _a1
func (_m *NamespaceDiscoveryRepository) CreateNamespace(_a0 context.Context, _a1 *namespace.Namespace) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *namespace.Namespace) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NamespaceDiscoveryRepository_CreateNamespace_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateNamespace'
type NamespaceDiscoveryRepository_CreateNamespace_Call struct {
	*mock.Call
}

// CreateNamespace is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *namespace.Namespace
func (_e *NamespaceDiscoveryRepository_Expecter) CreateNamespace(_a0 interface{}, _a1 interface{}) *NamespaceDiscoveryRepository_CreateNamespace_Call {
	return &NamespaceDiscoveryRepository_CreateNamespace_Call{Call: _e.mock.On("CreateNamespace", _a0, _a1)}
}

func (_c *NamespaceDiscoveryRepository_CreateNamespace_Call) Run(run func(_a0 context.Context, _a1 *namespace.Namespace)) *NamespaceDiscoveryRepository_CreateNamespace_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*namespace.Namespace))
	})
	return _c
}

func (_c *NamespaceDiscoveryRepository_CreateNamespace_Call) Return(_a0 error) *NamespaceDiscoveryRepository_CreateNamespace_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NamespaceDiscoveryRepository_CreateNamespace_Call) RunAndReturn(run func(context.Context, *namespace.Namespace) error) *NamespaceDiscoveryRepository_CreateNamespace_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewNamespaceDiscoveryRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewNamespaceDiscoveryRepository creates a new instance of NamespaceDiscoveryRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewNamespaceDiscoveryRepository(t mockConstructorTestingTNewNamespaceDiscoveryRepository) *NamespaceDiscoveryRepository {
	mock := &NamespaceDiscoveryRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}