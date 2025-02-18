// Code generated by mockery v2.50.0. DO NOT EDIT.

package getlink

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MocklinksService is an autogenerated mock type for the linksService type
type MocklinksService struct {
	mock.Mock
}

type MocklinksService_Expecter struct {
	mock *mock.Mock
}

func (_m *MocklinksService) EXPECT() *MocklinksService_Expecter {
	return &MocklinksService_Expecter{mock: &_m.Mock}
}

// Get provides a mock function with given fields: ctx, shortLink
func (_m *MocklinksService) Get(ctx context.Context, shortLink string) (string, error) {
	ret := _m.Called(ctx, shortLink)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, shortLink)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, shortLink)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, shortLink)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MocklinksService_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type MocklinksService_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - shortLink string
func (_e *MocklinksService_Expecter) Get(ctx interface{}, shortLink interface{}) *MocklinksService_Get_Call {
	return &MocklinksService_Get_Call{Call: _e.mock.On("Get", ctx, shortLink)}
}

func (_c *MocklinksService_Get_Call) Run(run func(ctx context.Context, shortLink string)) *MocklinksService_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MocklinksService_Get_Call) Return(_a0 string, _a1 error) *MocklinksService_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MocklinksService_Get_Call) RunAndReturn(run func(context.Context, string) (string, error)) *MocklinksService_Get_Call {
	_c.Call.Return(run)
	return _c
}

// NewMocklinksService creates a new instance of MocklinksService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMocklinksService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MocklinksService {
	mock := &MocklinksService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
