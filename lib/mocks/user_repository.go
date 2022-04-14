// Code generated by mockery v2.10.4. DO NOT EDIT.

package mocks

import (
	context "context"

	user "github.com/odpf/columbus/user"
	mock "github.com/stretchr/testify/mock"
)

// UserRepository is an autogenerated mock type for the Repository type
type UserRepository struct {
	mock.Mock
}

type UserRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *UserRepository) EXPECT() *UserRepository_Expecter {
	return &UserRepository_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: ctx, u
func (_m *UserRepository) Create(ctx context.Context, u *user.User) (string, error) {
	ret := _m.Called(ctx, u)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, *user.User) string); ok {
		r0 = rf(ctx, u)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *user.User) error); ok {
		r1 = rf(ctx, u)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UserRepository_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type UserRepository_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//  - ctx context.Context
//  - u *user.User
func (_e *UserRepository_Expecter) Create(ctx interface{}, u interface{}) *UserRepository_Create_Call {
	return &UserRepository_Create_Call{Call: _e.mock.On("Create", ctx, u)}
}

func (_c *UserRepository_Create_Call) Run(run func(ctx context.Context, u *user.User)) *UserRepository_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*user.User))
	})
	return _c
}

func (_c *UserRepository_Create_Call) Return(_a0 string, _a1 error) *UserRepository_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetID provides a mock function with given fields: ctx, email
func (_m *UserRepository) GetID(ctx context.Context, email string) (string, error) {
	ret := _m.Called(ctx, email)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, email)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, email)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UserRepository_GetID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetID'
type UserRepository_GetID_Call struct {
	*mock.Call
}

// GetID is a helper method to define mock.On call
//  - ctx context.Context
//  - email string
func (_e *UserRepository_Expecter) GetID(ctx interface{}, email interface{}) *UserRepository_GetID_Call {
	return &UserRepository_GetID_Call{Call: _e.mock.On("GetID", ctx, email)}
}

func (_c *UserRepository_GetID_Call) Run(run func(ctx context.Context, email string)) *UserRepository_GetID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *UserRepository_GetID_Call) Return(_a0 string, _a1 error) *UserRepository_GetID_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}
