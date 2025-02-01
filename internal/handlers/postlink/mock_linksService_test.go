// Code generated by mockery v2.50.0. DO NOT EDIT.

package postlink

import mock "github.com/stretchr/testify/mock"

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

// Add provides a mock function with given fields: long
func (_m *MocklinksService) Add(long string) (string, error) {
	ret := _m.Called(long)

	if len(ret) == 0 {
		panic("no return value specified for Add")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (string, error)); ok {
		return rf(long)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(long)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(long)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MocklinksService_Add_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Add'
type MocklinksService_Add_Call struct {
	*mock.Call
}

// Add is a helper method to define mock.On call
//   - long string
func (_e *MocklinksService_Expecter) Add(long interface{}) *MocklinksService_Add_Call {
	return &MocklinksService_Add_Call{Call: _e.mock.On("Add", long)}
}

func (_c *MocklinksService_Add_Call) Run(run func(long string)) *MocklinksService_Add_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MocklinksService_Add_Call) Return(_a0 string, _a1 error) *MocklinksService_Add_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MocklinksService_Add_Call) RunAndReturn(run func(string) (string, error)) *MocklinksService_Add_Call {
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
