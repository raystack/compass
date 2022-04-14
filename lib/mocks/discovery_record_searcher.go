// Code generated by mockery v2.10.4. DO NOT EDIT.

package mocks

import (
	context "context"

	discovery "github.com/odpf/columbus/discovery"
	mock "github.com/stretchr/testify/mock"
)

// DiscoveryRecordSearcher is an autogenerated mock type for the RecordSearcher type
type DiscoveryRecordSearcher struct {
	mock.Mock
}

type DiscoveryRecordSearcher_Expecter struct {
	mock *mock.Mock
}

func (_m *DiscoveryRecordSearcher) EXPECT() *DiscoveryRecordSearcher_Expecter {
	return &DiscoveryRecordSearcher_Expecter{mock: &_m.Mock}
}

// Search provides a mock function with given fields: ctx, cfg
func (_m *DiscoveryRecordSearcher) Search(ctx context.Context, cfg discovery.SearchConfig) ([]discovery.SearchResult, error) {
	ret := _m.Called(ctx, cfg)

	var r0 []discovery.SearchResult
	if rf, ok := ret.Get(0).(func(context.Context, discovery.SearchConfig) []discovery.SearchResult); ok {
		r0 = rf(ctx, cfg)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]discovery.SearchResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, discovery.SearchConfig) error); ok {
		r1 = rf(ctx, cfg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DiscoveryRecordSearcher_Search_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Search'
type DiscoveryRecordSearcher_Search_Call struct {
	*mock.Call
}

// Search is a helper method to define mock.On call
//  - ctx context.Context
//  - cfg discovery.SearchConfig
func (_e *DiscoveryRecordSearcher_Expecter) Search(ctx interface{}, cfg interface{}) *DiscoveryRecordSearcher_Search_Call {
	return &DiscoveryRecordSearcher_Search_Call{Call: _e.mock.On("Search", ctx, cfg)}
}

func (_c *DiscoveryRecordSearcher_Search_Call) Run(run func(ctx context.Context, cfg discovery.SearchConfig)) *DiscoveryRecordSearcher_Search_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(discovery.SearchConfig))
	})
	return _c
}

func (_c *DiscoveryRecordSearcher_Search_Call) Return(results []discovery.SearchResult, err error) *DiscoveryRecordSearcher_Search_Call {
	_c.Call.Return(results, err)
	return _c
}

// Suggest provides a mock function with given fields: ctx, cfg
func (_m *DiscoveryRecordSearcher) Suggest(ctx context.Context, cfg discovery.SearchConfig) ([]string, error) {
	ret := _m.Called(ctx, cfg)

	var r0 []string
	if rf, ok := ret.Get(0).(func(context.Context, discovery.SearchConfig) []string); ok {
		r0 = rf(ctx, cfg)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, discovery.SearchConfig) error); ok {
		r1 = rf(ctx, cfg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DiscoveryRecordSearcher_Suggest_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Suggest'
type DiscoveryRecordSearcher_Suggest_Call struct {
	*mock.Call
}

// Suggest is a helper method to define mock.On call
//  - ctx context.Context
//  - cfg discovery.SearchConfig
func (_e *DiscoveryRecordSearcher_Expecter) Suggest(ctx interface{}, cfg interface{}) *DiscoveryRecordSearcher_Suggest_Call {
	return &DiscoveryRecordSearcher_Suggest_Call{Call: _e.mock.On("Suggest", ctx, cfg)}
}

func (_c *DiscoveryRecordSearcher_Suggest_Call) Run(run func(ctx context.Context, cfg discovery.SearchConfig)) *DiscoveryRecordSearcher_Suggest_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(discovery.SearchConfig))
	})
	return _c
}

func (_c *DiscoveryRecordSearcher_Suggest_Call) Return(suggestions []string, err error) *DiscoveryRecordSearcher_Suggest_Call {
	_c.Call.Return(suggestions, err)
	return _c
}
