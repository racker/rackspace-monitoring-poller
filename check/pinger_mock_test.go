// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/racker/rackspace-monitoring-poller/check (interfaces: Pinger)

package check

import (
	gomock "github.com/golang/mock/gomock"
	go_ping "github.com/sparrc/go-ping"
	time "time"
)

// Mock of Pinger interface
type MockPinger struct {
	ctrl     *gomock.Controller
	recorder *_MockPingerRecorder
}

// Recorder for MockPinger (not exported)
type _MockPingerRecorder struct {
	mock *MockPinger
}

func NewMockPinger(ctrl *gomock.Controller) *MockPinger {
	mock := &MockPinger{ctrl: ctrl}
	mock.recorder = &_MockPingerRecorder{mock}
	return mock
}

func (_m *MockPinger) EXPECT() *_MockPingerRecorder {
	return _m.recorder
}

func (_m *MockPinger) Count() int {
	ret := _m.ctrl.Call(_m, "Count")
	ret0, _ := ret[0].(int)
	return ret0
}

func (_mr *_MockPingerRecorder) Count() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Count")
}

func (_m *MockPinger) Run() {
	_m.ctrl.Call(_m, "Run")
}

func (_mr *_MockPingerRecorder) Run() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Run")
}

func (_m *MockPinger) SetCount(_param0 int) {
	_m.ctrl.Call(_m, "SetCount", _param0)
}

func (_mr *_MockPingerRecorder) SetCount(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetCount", arg0)
}

func (_m *MockPinger) SetOnRecv(_param0 func(*go_ping.Packet)) {
	_m.ctrl.Call(_m, "SetOnRecv", _param0)
}

func (_mr *_MockPingerRecorder) SetOnRecv(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetOnRecv", arg0)
}

func (_m *MockPinger) SetTimeout(_param0 time.Duration) {
	_m.ctrl.Call(_m, "SetTimeout", _param0)
}

func (_mr *_MockPingerRecorder) SetTimeout(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetTimeout", arg0)
}

func (_m *MockPinger) Statistics() *go_ping.Statistics {
	ret := _m.ctrl.Call(_m, "Statistics")
	ret0, _ := ret[0].(*go_ping.Statistics)
	return ret0
}

func (_mr *_MockPingerRecorder) Statistics() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Statistics")
}

func (_m *MockPinger) Timeout() time.Duration {
	ret := _m.ctrl.Call(_m, "Timeout")
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

func (_mr *_MockPingerRecorder) Timeout() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Timeout")
}
