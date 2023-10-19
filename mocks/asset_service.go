// Code generated by mockery v2.26.1. DO NOT EDIT.

package mocks

import (
	context "context"

	asset "github.com/goto/compass/core/asset"

	mock "github.com/stretchr/testify/mock"
)

// AssetService is an autogenerated mock type for the AssetService type
type AssetService struct {
	mock.Mock
}

type AssetService_Expecter struct {
	mock *mock.Mock
}

func (_m *AssetService) EXPECT() *AssetService_Expecter {
	return &AssetService_Expecter{mock: &_m.Mock}
}

// AddProbe provides a mock function with given fields: ctx, assetURN, probe
func (_m *AssetService) AddProbe(ctx context.Context, assetURN string, probe *asset.Probe) error {
	ret := _m.Called(ctx, assetURN, probe)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *asset.Probe) error); ok {
		r0 = rf(ctx, assetURN, probe)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AssetService_AddProbe_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddProbe'
type AssetService_AddProbe_Call struct {
	*mock.Call
}

// AddProbe is a helper method to define mock.On call
//   - ctx context.Context
//   - assetURN string
//   - probe *asset.Probe
func (_e *AssetService_Expecter) AddProbe(ctx interface{}, assetURN interface{}, probe interface{}) *AssetService_AddProbe_Call {
	return &AssetService_AddProbe_Call{Call: _e.mock.On("AddProbe", ctx, assetURN, probe)}
}

func (_c *AssetService_AddProbe_Call) Run(run func(ctx context.Context, assetURN string, probe *asset.Probe)) *AssetService_AddProbe_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(*asset.Probe))
	})
	return _c
}

func (_c *AssetService_AddProbe_Call) Return(_a0 error) *AssetService_AddProbe_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AssetService_AddProbe_Call) RunAndReturn(run func(context.Context, string, *asset.Probe) error) *AssetService_AddProbe_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteAsset provides a mock function with given fields: ctx, id
func (_m *AssetService) DeleteAsset(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AssetService_DeleteAsset_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteAsset'
type AssetService_DeleteAsset_Call struct {
	*mock.Call
}

// DeleteAsset is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *AssetService_Expecter) DeleteAsset(ctx interface{}, id interface{}) *AssetService_DeleteAsset_Call {
	return &AssetService_DeleteAsset_Call{Call: _e.mock.On("DeleteAsset", ctx, id)}
}

func (_c *AssetService_DeleteAsset_Call) Run(run func(ctx context.Context, id string)) *AssetService_DeleteAsset_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *AssetService_DeleteAsset_Call) Return(_a0 error) *AssetService_DeleteAsset_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AssetService_DeleteAsset_Call) RunAndReturn(run func(context.Context, string) error) *AssetService_DeleteAsset_Call {
	_c.Call.Return(run)
	return _c
}

// GetAllAssets provides a mock function with given fields: ctx, flt, withTotal
func (_m *AssetService) GetAllAssets(ctx context.Context, flt asset.Filter, withTotal bool) ([]asset.Asset, uint32, error) {
	ret := _m.Called(ctx, flt, withTotal)

	var r0 []asset.Asset
	var r1 uint32
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, asset.Filter, bool) ([]asset.Asset, uint32, error)); ok {
		return rf(ctx, flt, withTotal)
	}
	if rf, ok := ret.Get(0).(func(context.Context, asset.Filter, bool) []asset.Asset); ok {
		r0 = rf(ctx, flt, withTotal)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]asset.Asset)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, asset.Filter, bool) uint32); ok {
		r1 = rf(ctx, flt, withTotal)
	} else {
		r1 = ret.Get(1).(uint32)
	}

	if rf, ok := ret.Get(2).(func(context.Context, asset.Filter, bool) error); ok {
		r2 = rf(ctx, flt, withTotal)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// AssetService_GetAllAssets_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAllAssets'
type AssetService_GetAllAssets_Call struct {
	*mock.Call
}

// GetAllAssets is a helper method to define mock.On call
//   - ctx context.Context
//   - flt asset.Filter
//   - withTotal bool
func (_e *AssetService_Expecter) GetAllAssets(ctx interface{}, flt interface{}, withTotal interface{}) *AssetService_GetAllAssets_Call {
	return &AssetService_GetAllAssets_Call{Call: _e.mock.On("GetAllAssets", ctx, flt, withTotal)}
}

func (_c *AssetService_GetAllAssets_Call) Run(run func(ctx context.Context, flt asset.Filter, withTotal bool)) *AssetService_GetAllAssets_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(asset.Filter), args[2].(bool))
	})
	return _c
}

func (_c *AssetService_GetAllAssets_Call) Return(_a0 []asset.Asset, _a1 uint32, _a2 error) *AssetService_GetAllAssets_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *AssetService_GetAllAssets_Call) RunAndReturn(run func(context.Context, asset.Filter, bool) ([]asset.Asset, uint32, error)) *AssetService_GetAllAssets_Call {
	_c.Call.Return(run)
	return _c
}

// GetAssetByID provides a mock function with given fields: ctx, id
func (_m *AssetService) GetAssetByID(ctx context.Context, id string) (asset.Asset, error) {
	ret := _m.Called(ctx, id)

	var r0 asset.Asset
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (asset.Asset, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) asset.Asset); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(asset.Asset)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetService_GetAssetByID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAssetByID'
type AssetService_GetAssetByID_Call struct {
	*mock.Call
}

// GetAssetByID is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *AssetService_Expecter) GetAssetByID(ctx interface{}, id interface{}) *AssetService_GetAssetByID_Call {
	return &AssetService_GetAssetByID_Call{Call: _e.mock.On("GetAssetByID", ctx, id)}
}

func (_c *AssetService_GetAssetByID_Call) Run(run func(ctx context.Context, id string)) *AssetService_GetAssetByID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *AssetService_GetAssetByID_Call) Return(_a0 asset.Asset, _a1 error) *AssetService_GetAssetByID_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AssetService_GetAssetByID_Call) RunAndReturn(run func(context.Context, string) (asset.Asset, error)) *AssetService_GetAssetByID_Call {
	_c.Call.Return(run)
	return _c
}

// GetAssetByVersion provides a mock function with given fields: ctx, id, version
func (_m *AssetService) GetAssetByVersion(ctx context.Context, id string, version string) (asset.Asset, error) {
	ret := _m.Called(ctx, id, version)

	var r0 asset.Asset
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (asset.Asset, error)); ok {
		return rf(ctx, id, version)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) asset.Asset); ok {
		r0 = rf(ctx, id, version)
	} else {
		r0 = ret.Get(0).(asset.Asset)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, id, version)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetService_GetAssetByVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAssetByVersion'
type AssetService_GetAssetByVersion_Call struct {
	*mock.Call
}

// GetAssetByVersion is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
//   - version string
func (_e *AssetService_Expecter) GetAssetByVersion(ctx interface{}, id interface{}, version interface{}) *AssetService_GetAssetByVersion_Call {
	return &AssetService_GetAssetByVersion_Call{Call: _e.mock.On("GetAssetByVersion", ctx, id, version)}
}

func (_c *AssetService_GetAssetByVersion_Call) Run(run func(ctx context.Context, id string, version string)) *AssetService_GetAssetByVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *AssetService_GetAssetByVersion_Call) Return(_a0 asset.Asset, _a1 error) *AssetService_GetAssetByVersion_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AssetService_GetAssetByVersion_Call) RunAndReturn(run func(context.Context, string, string) (asset.Asset, error)) *AssetService_GetAssetByVersion_Call {
	_c.Call.Return(run)
	return _c
}

// GetAssetVersionHistory provides a mock function with given fields: ctx, flt, id
func (_m *AssetService) GetAssetVersionHistory(ctx context.Context, flt asset.Filter, id string) ([]asset.Asset, error) {
	ret := _m.Called(ctx, flt, id)

	var r0 []asset.Asset
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, asset.Filter, string) ([]asset.Asset, error)); ok {
		return rf(ctx, flt, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, asset.Filter, string) []asset.Asset); ok {
		r0 = rf(ctx, flt, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]asset.Asset)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, asset.Filter, string) error); ok {
		r1 = rf(ctx, flt, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetService_GetAssetVersionHistory_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAssetVersionHistory'
type AssetService_GetAssetVersionHistory_Call struct {
	*mock.Call
}

// GetAssetVersionHistory is a helper method to define mock.On call
//   - ctx context.Context
//   - flt asset.Filter
//   - id string
func (_e *AssetService_Expecter) GetAssetVersionHistory(ctx interface{}, flt interface{}, id interface{}) *AssetService_GetAssetVersionHistory_Call {
	return &AssetService_GetAssetVersionHistory_Call{Call: _e.mock.On("GetAssetVersionHistory", ctx, flt, id)}
}

func (_c *AssetService_GetAssetVersionHistory_Call) Run(run func(ctx context.Context, flt asset.Filter, id string)) *AssetService_GetAssetVersionHistory_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(asset.Filter), args[2].(string))
	})
	return _c
}

func (_c *AssetService_GetAssetVersionHistory_Call) Return(_a0 []asset.Asset, _a1 error) *AssetService_GetAssetVersionHistory_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AssetService_GetAssetVersionHistory_Call) RunAndReturn(run func(context.Context, asset.Filter, string) ([]asset.Asset, error)) *AssetService_GetAssetVersionHistory_Call {
	_c.Call.Return(run)
	return _c
}

// GetLineage provides a mock function with given fields: ctx, urn, query
func (_m *AssetService) GetLineage(ctx context.Context, urn string, query asset.LineageQuery) (asset.Lineage, error) {
	ret := _m.Called(ctx, urn, query)

	var r0 asset.Lineage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, asset.LineageQuery) (asset.Lineage, error)); ok {
		return rf(ctx, urn, query)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, asset.LineageQuery) asset.Lineage); ok {
		r0 = rf(ctx, urn, query)
	} else {
		r0 = ret.Get(0).(asset.Lineage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, asset.LineageQuery) error); ok {
		r1 = rf(ctx, urn, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetService_GetLineage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLineage'
type AssetService_GetLineage_Call struct {
	*mock.Call
}

// GetLineage is a helper method to define mock.On call
//   - ctx context.Context
//   - urn string
//   - query asset.LineageQuery
func (_e *AssetService_Expecter) GetLineage(ctx interface{}, urn interface{}, query interface{}) *AssetService_GetLineage_Call {
	return &AssetService_GetLineage_Call{Call: _e.mock.On("GetLineage", ctx, urn, query)}
}

func (_c *AssetService_GetLineage_Call) Run(run func(ctx context.Context, urn string, query asset.LineageQuery)) *AssetService_GetLineage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(asset.LineageQuery))
	})
	return _c
}

func (_c *AssetService_GetLineage_Call) Return(_a0 asset.Lineage, _a1 error) *AssetService_GetLineage_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AssetService_GetLineage_Call) RunAndReturn(run func(context.Context, string, asset.LineageQuery) (asset.Lineage, error)) *AssetService_GetLineage_Call {
	_c.Call.Return(run)
	return _c
}

// GetTypes provides a mock function with given fields: ctx, flt
func (_m *AssetService) GetTypes(ctx context.Context, flt asset.Filter) (map[asset.Type]int, error) {
	ret := _m.Called(ctx, flt)

	var r0 map[asset.Type]int
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, asset.Filter) (map[asset.Type]int, error)); ok {
		return rf(ctx, flt)
	}
	if rf, ok := ret.Get(0).(func(context.Context, asset.Filter) map[asset.Type]int); ok {
		r0 = rf(ctx, flt)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[asset.Type]int)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, asset.Filter) error); ok {
		r1 = rf(ctx, flt)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetService_GetTypes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTypes'
type AssetService_GetTypes_Call struct {
	*mock.Call
}

// GetTypes is a helper method to define mock.On call
//   - ctx context.Context
//   - flt asset.Filter
func (_e *AssetService_Expecter) GetTypes(ctx interface{}, flt interface{}) *AssetService_GetTypes_Call {
	return &AssetService_GetTypes_Call{Call: _e.mock.On("GetTypes", ctx, flt)}
}

func (_c *AssetService_GetTypes_Call) Run(run func(ctx context.Context, flt asset.Filter)) *AssetService_GetTypes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(asset.Filter))
	})
	return _c
}

func (_c *AssetService_GetTypes_Call) Return(_a0 map[asset.Type]int, _a1 error) *AssetService_GetTypes_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AssetService_GetTypes_Call) RunAndReturn(run func(context.Context, asset.Filter) (map[asset.Type]int, error)) *AssetService_GetTypes_Call {
	_c.Call.Return(run)
	return _c
}

// GroupAssets provides a mock function with given fields: ctx, cfg
func (_m *AssetService) GroupAssets(ctx context.Context, cfg asset.GroupConfig) ([]asset.GroupResult, error) {
	ret := _m.Called(ctx, cfg)

	var r0 []asset.GroupResult
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, asset.GroupConfig) ([]asset.GroupResult, error)); ok {
		return rf(ctx, cfg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, asset.GroupConfig) []asset.GroupResult); ok {
		r0 = rf(ctx, cfg)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]asset.GroupResult)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, asset.GroupConfig) error); ok {
		r1 = rf(ctx, cfg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetService_GroupAssets_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GroupAssets'
type AssetService_GroupAssets_Call struct {
	*mock.Call
}

// GroupAssets is a helper method to define mock.On call
//   - ctx context.Context
//   - cfg asset.GroupConfig
func (_e *AssetService_Expecter) GroupAssets(ctx interface{}, cfg interface{}) *AssetService_GroupAssets_Call {
	return &AssetService_GroupAssets_Call{Call: _e.mock.On("GroupAssets", ctx, cfg)}
}

func (_c *AssetService_GroupAssets_Call) Run(run func(ctx context.Context, cfg asset.GroupConfig)) *AssetService_GroupAssets_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(asset.GroupConfig))
	})
	return _c
}

func (_c *AssetService_GroupAssets_Call) Return(results []asset.GroupResult, err error) *AssetService_GroupAssets_Call {
	_c.Call.Return(results, err)
	return _c
}

func (_c *AssetService_GroupAssets_Call) RunAndReturn(run func(context.Context, asset.GroupConfig) ([]asset.GroupResult, error)) *AssetService_GroupAssets_Call {
	_c.Call.Return(run)
	return _c
}

// SearchAssets provides a mock function with given fields: ctx, cfg
func (_m *AssetService) SearchAssets(ctx context.Context, cfg asset.SearchConfig) ([]asset.SearchResult, error) {
	ret := _m.Called(ctx, cfg)

	var r0 []asset.SearchResult
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, asset.SearchConfig) ([]asset.SearchResult, error)); ok {
		return rf(ctx, cfg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, asset.SearchConfig) []asset.SearchResult); ok {
		r0 = rf(ctx, cfg)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]asset.SearchResult)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, asset.SearchConfig) error); ok {
		r1 = rf(ctx, cfg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetService_SearchAssets_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SearchAssets'
type AssetService_SearchAssets_Call struct {
	*mock.Call
}

// SearchAssets is a helper method to define mock.On call
//   - ctx context.Context
//   - cfg asset.SearchConfig
func (_e *AssetService_Expecter) SearchAssets(ctx interface{}, cfg interface{}) *AssetService_SearchAssets_Call {
	return &AssetService_SearchAssets_Call{Call: _e.mock.On("SearchAssets", ctx, cfg)}
}

func (_c *AssetService_SearchAssets_Call) Run(run func(ctx context.Context, cfg asset.SearchConfig)) *AssetService_SearchAssets_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(asset.SearchConfig))
	})
	return _c
}

func (_c *AssetService_SearchAssets_Call) Return(results []asset.SearchResult, err error) *AssetService_SearchAssets_Call {
	_c.Call.Return(results, err)
	return _c
}

func (_c *AssetService_SearchAssets_Call) RunAndReturn(run func(context.Context, asset.SearchConfig) ([]asset.SearchResult, error)) *AssetService_SearchAssets_Call {
	_c.Call.Return(run)
	return _c
}

// SuggestAssets provides a mock function with given fields: ctx, cfg
func (_m *AssetService) SuggestAssets(ctx context.Context, cfg asset.SearchConfig) ([]string, error) {
	ret := _m.Called(ctx, cfg)

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, asset.SearchConfig) ([]string, error)); ok {
		return rf(ctx, cfg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, asset.SearchConfig) []string); ok {
		r0 = rf(ctx, cfg)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, asset.SearchConfig) error); ok {
		r1 = rf(ctx, cfg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetService_SuggestAssets_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SuggestAssets'
type AssetService_SuggestAssets_Call struct {
	*mock.Call
}

// SuggestAssets is a helper method to define mock.On call
//   - ctx context.Context
//   - cfg asset.SearchConfig
func (_e *AssetService_Expecter) SuggestAssets(ctx interface{}, cfg interface{}) *AssetService_SuggestAssets_Call {
	return &AssetService_SuggestAssets_Call{Call: _e.mock.On("SuggestAssets", ctx, cfg)}
}

func (_c *AssetService_SuggestAssets_Call) Run(run func(ctx context.Context, cfg asset.SearchConfig)) *AssetService_SuggestAssets_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(asset.SearchConfig))
	})
	return _c
}

func (_c *AssetService_SuggestAssets_Call) Return(suggestions []string, err error) *AssetService_SuggestAssets_Call {
	_c.Call.Return(suggestions, err)
	return _c
}

func (_c *AssetService_SuggestAssets_Call) RunAndReturn(run func(context.Context, asset.SearchConfig) ([]string, error)) *AssetService_SuggestAssets_Call {
	_c.Call.Return(run)
	return _c
}

// UpsertAsset provides a mock function with given fields: ctx, ast, upstreams, downstreams
func (_m *AssetService) UpsertAsset(ctx context.Context, ast *asset.Asset, upstreams []string, downstreams []string) (string, error) {
	ret := _m.Called(ctx, ast, upstreams, downstreams)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *asset.Asset, []string, []string) (string, error)); ok {
		return rf(ctx, ast, upstreams, downstreams)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *asset.Asset, []string, []string) string); ok {
		r0 = rf(ctx, ast, upstreams, downstreams)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *asset.Asset, []string, []string) error); ok {
		r1 = rf(ctx, ast, upstreams, downstreams)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetService_UpsertAsset_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpsertAsset'
type AssetService_UpsertAsset_Call struct {
	*mock.Call
}

// UpsertAsset is a helper method to define mock.On call
//   - ctx context.Context
//   - ast *asset.Asset
//   - upstreams []string
//   - downstreams []string
func (_e *AssetService_Expecter) UpsertAsset(ctx interface{}, ast interface{}, upstreams interface{}, downstreams interface{}) *AssetService_UpsertAsset_Call {
	return &AssetService_UpsertAsset_Call{Call: _e.mock.On("UpsertAsset", ctx, ast, upstreams, downstreams)}
}

func (_c *AssetService_UpsertAsset_Call) Run(run func(ctx context.Context, ast *asset.Asset, upstreams []string, downstreams []string)) *AssetService_UpsertAsset_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*asset.Asset), args[2].([]string), args[3].([]string))
	})
	return _c
}

func (_c *AssetService_UpsertAsset_Call) Return(_a0 string, _a1 error) *AssetService_UpsertAsset_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AssetService_UpsertAsset_Call) RunAndReturn(run func(context.Context, *asset.Asset, []string, []string) (string, error)) *AssetService_UpsertAsset_Call {
	_c.Call.Return(run)
	return _c
}

// UpsertAssetWithoutLineage provides a mock function with given fields: ctx, ast
func (_m *AssetService) UpsertAssetWithoutLineage(ctx context.Context, ast *asset.Asset) (string, error) {
	ret := _m.Called(ctx, ast)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *asset.Asset) (string, error)); ok {
		return rf(ctx, ast)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *asset.Asset) string); ok {
		r0 = rf(ctx, ast)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *asset.Asset) error); ok {
		r1 = rf(ctx, ast)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetService_UpsertAssetWithoutLineage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpsertAssetWithoutLineage'
type AssetService_UpsertAssetWithoutLineage_Call struct {
	*mock.Call
}

// UpsertAssetWithoutLineage is a helper method to define mock.On call
//   - ctx context.Context
//   - ast *asset.Asset
func (_e *AssetService_Expecter) UpsertAssetWithoutLineage(ctx interface{}, ast interface{}) *AssetService_UpsertAssetWithoutLineage_Call {
	return &AssetService_UpsertAssetWithoutLineage_Call{Call: _e.mock.On("UpsertAssetWithoutLineage", ctx, ast)}
}

func (_c *AssetService_UpsertAssetWithoutLineage_Call) Run(run func(ctx context.Context, ast *asset.Asset)) *AssetService_UpsertAssetWithoutLineage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*asset.Asset))
	})
	return _c
}

func (_c *AssetService_UpsertAssetWithoutLineage_Call) Return(_a0 string, _a1 error) *AssetService_UpsertAssetWithoutLineage_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AssetService_UpsertAssetWithoutLineage_Call) RunAndReturn(run func(context.Context, *asset.Asset) (string, error)) *AssetService_UpsertAssetWithoutLineage_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewAssetService interface {
	mock.TestingT
	Cleanup(func())
}

// NewAssetService creates a new instance of AssetService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewAssetService(t mockConstructorTestingTNewAssetService) *AssetService {
	mock := &AssetService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}