// Code generated by mockery 2.7.5. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// BrowserHandler is an autogenerated mock type for the BrowserHandler type
type BrowserHandler struct {
	mock.Mock
}

// Evaluate provides a mock function with given fields: ctx, script
func (_m *BrowserHandler) Evaluate(ctx context.Context, script string) ([]byte, error) {
	ret := _m.Called(ctx, script)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(context.Context, string) []byte); ok {
		r0 = rf(ctx, script)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, script)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Navigate provides a mock function with given fields: ctx, urlstr, width, height
func (_m *BrowserHandler) Navigate(ctx context.Context, urlstr string, width int64, height int64) error {
	ret := _m.Called(ctx, urlstr, width, height)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, int64, int64) error); ok {
		r0 = rf(ctx, urlstr, width, height)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewContext provides a mock function with given fields: parent
func (_m *BrowserHandler) NewContext(parent context.Context) (context.Context, context.CancelFunc) {
	ret := _m.Called(parent)

	var r0 context.Context
	if rf, ok := ret.Get(0).(func(context.Context) context.Context); ok {
		r0 = rf(parent)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(context.Context)
		}
	}

	var r1 context.CancelFunc
	if rf, ok := ret.Get(1).(func(context.Context) context.CancelFunc); ok {
		r1 = rf(parent)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(context.CancelFunc)
		}
	}

	return r0, r1
}

// Screenshot provides a mock function with given fields: ctx
func (_m *BrowserHandler) Screenshot(ctx context.Context) ([]byte, error) {
	ret := _m.Called(ctx)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(context.Context) []byte); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
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