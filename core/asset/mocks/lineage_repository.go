// Code generated by mockery v2.12.2. DO NOT EDIT.

package mocks

import (
	context "context"

	asset "github.com/odpf/compass/core/asset"

	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// LineageRepository is an autogenerated mock type for the LineageRepository type
type LineageRepository struct {
	mock.Mock
}

type LineageRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *LineageRepository) EXPECT() *LineageRepository_Expecter {
	return &LineageRepository_Expecter{mock: &_m.Mock}
}

// GetGraph provides a mock function with given fields: ctx, node
func (_m *LineageRepository) GetGraph(ctx context.Context, node asset.LineageNode) (asset.LineageGraph, error) {
	ret := _m.Called(ctx, node)

	var r0 asset.LineageGraph
	if rf, ok := ret.Get(0).(func(context.Context, asset.LineageNode) asset.LineageGraph); ok {
		r0 = rf(ctx, node)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(asset.LineageGraph)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, asset.LineageNode) error); ok {
		r1 = rf(ctx, node)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LineageRepository_GetGraph_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetGraph'
type LineageRepository_GetGraph_Call struct {
	*mock.Call
}

// GetGraph is a helper method to define mock.On call
//  - ctx context.Context
//  - node asset.LineageNode
func (_e *LineageRepository_Expecter) GetGraph(ctx interface{}, node interface{}) *LineageRepository_GetGraph_Call {
	return &LineageRepository_GetGraph_Call{Call: _e.mock.On("GetGraph", ctx, node)}
}

func (_c *LineageRepository_GetGraph_Call) Run(run func(ctx context.Context, node asset.LineageNode)) *LineageRepository_GetGraph_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(asset.LineageNode))
	})
	return _c
}

func (_c *LineageRepository_GetGraph_Call) Return(_a0 asset.LineageGraph, _a1 error) *LineageRepository_GetGraph_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Upsert provides a mock function with given fields: ctx, node, upstreams, downstreams
func (_m *LineageRepository) Upsert(ctx context.Context, node asset.LineageNode, upstreams []asset.LineageNode, downstreams []asset.LineageNode) error {
	ret := _m.Called(ctx, node, upstreams, downstreams)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, asset.LineageNode, []asset.LineageNode, []asset.LineageNode) error); ok {
		r0 = rf(ctx, node, upstreams, downstreams)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// LineageRepository_Upsert_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Upsert'
type LineageRepository_Upsert_Call struct {
	*mock.Call
}

// Upsert is a helper method to define mock.On call
//  - ctx context.Context
//  - node asset.LineageNode
//  - upstreams []asset.LineageNode
//  - downstreams []asset.LineageNode
func (_e *LineageRepository_Expecter) Upsert(ctx interface{}, node interface{}, upstreams interface{}, downstreams interface{}) *LineageRepository_Upsert_Call {
	return &LineageRepository_Upsert_Call{Call: _e.mock.On("Upsert", ctx, node, upstreams, downstreams)}
}

func (_c *LineageRepository_Upsert_Call) Run(run func(ctx context.Context, node asset.LineageNode, upstreams []asset.LineageNode, downstreams []asset.LineageNode)) *LineageRepository_Upsert_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(asset.LineageNode), args[2].([]asset.LineageNode), args[3].([]asset.LineageNode))
	})
	return _c
}

func (_c *LineageRepository_Upsert_Call) Return(_a0 error) *LineageRepository_Upsert_Call {
	_c.Call.Return(_a0)
	return _c
}

// NewLineageRepository creates a new instance of LineageRepository. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewLineageRepository(t testing.TB) *LineageRepository {
	mock := &LineageRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
