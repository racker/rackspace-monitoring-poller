// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/racker/rackspace-monitoring-poller/check (interfaces: Pinger)

package check_test

import (
	gomock "github.com/golang/mock/gomock"
	check "github.com/racker/rackspace-monitoring-poller/check"
	reflect "reflect"
	time "time"
)

// MockPinger is a mock of Pinger interface
type MockPinger struct {
	ctrl     *gomock.Controller
	recorder *MockPingerMockRecorder
}

// MockPingerMockRecorder is the mock recorder for MockPinger
type MockPingerMockRecorder struct {
	mock *MockPinger
}

// NewMockPinger creates a new mock instance
func NewMockPinger(ctrl *gomock.Controller) *MockPinger {
	mock := &MockPinger{ctrl: ctrl}
	mock.recorder = &MockPingerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *MockPinger) EXPECT() *MockPingerMockRecorder {
	return _m.recorder
}

// Close mocks base method
func (_m *MockPinger) Close() {
	_m.ctrl.Call(_m, "Close")
}

// Close indicates an expected call of Close
func (_mr *MockPingerMockRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Close", reflect.TypeOf((*MockPinger)(nil).Close))
}

// Ping mocks base method
func (_m *MockPinger) Ping(_param0 int, _param1 time.Duration) check.PingResponse {
	ret := _m.ctrl.Call(_m, "Ping", _param0, _param1)
	ret0, _ := ret[0].(check.PingResponse)
	return ret0
}

// Ping indicates an expected call of Ping
func (_mr *MockPingerMockRecorder) Ping(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Ping", reflect.TypeOf((*MockPinger)(nil).Ping), arg0, arg1)
}
