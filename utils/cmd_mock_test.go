// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/racker/rackspace-monitoring-poller/utils (interfaces: Commander)

package utils

import (
	gomock "github.com/golang/mock/gomock"
	os "os"
)

// Mock of Commander interface
type MockCommander struct {
	ctrl     *gomock.Controller
	recorder *_MockCommanderRecorder
}

// Recorder for MockCommander (not exported)
type _MockCommanderRecorder struct {
	mock *MockCommander
}

func NewMockCommander(ctrl *gomock.Controller) *MockCommander {
	mock := &MockCommander{ctrl: ctrl}
	mock.recorder = &_MockCommanderRecorder{mock}
	return mock
}

func (_m *MockCommander) EXPECT() *_MockCommanderRecorder {
	return _m.recorder
}

func (_m *MockCommander) GetArgs() []string {
	ret := _m.ctrl.Call(_m, "GetArgs")
	ret0, _ := ret[0].([]string)
	return ret0
}

func (_mr *_MockCommanderRecorder) GetArgs() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetArgs")
}

func (_m *MockCommander) GetProcess() *os.Process {
	ret := _m.ctrl.Call(_m, "GetProcess")
	ret0, _ := ret[0].(*os.Process)
	return ret0
}

func (_mr *_MockCommanderRecorder) GetProcess() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetProcess")
}

func (_m *MockCommander) GetProcessState() *os.ProcessState {
	ret := _m.ctrl.Call(_m, "GetProcessState")
	ret0, _ := ret[0].(*os.ProcessState)
	return ret0
}

func (_mr *_MockCommanderRecorder) GetProcessState() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetProcessState")
}

func (_m *MockCommander) Start() error {
	ret := _m.ctrl.Call(_m, "Start")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockCommanderRecorder) Start() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Start")
}

func (_m *MockCommander) Wait() error {
	ret := _m.ctrl.Call(_m, "Wait")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockCommanderRecorder) Wait() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Wait")
}
