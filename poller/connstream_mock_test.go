// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/racker/rackspace-monitoring-poller/poller (interfaces: ConnectionStream)

package poller

import (
	context "context"

	gomock "github.com/golang/mock/gomock"
	check "github.com/racker/rackspace-monitoring-poller/check"
	config "github.com/racker/rackspace-monitoring-poller/config"
)

// Mock of ConnectionStream interface
type MockConnectionStream struct {
	ctrl     *gomock.Controller
	recorder *_MockConnectionStreamRecorder
}

// Recorder for MockConnectionStream (not exported)
type _MockConnectionStreamRecorder struct {
	mock *MockConnectionStream
}

func NewMockConnectionStream(ctrl *gomock.Controller) *MockConnectionStream {
	mock := &MockConnectionStream{ctrl: ctrl}
	mock.recorder = &_MockConnectionStreamRecorder{mock}
	return mock
}

func (_m *MockConnectionStream) EXPECT() *_MockConnectionStreamRecorder {
	return _m.recorder
}

func (_m *MockConnectionStream) Connect() {
	_m.ctrl.Call(_m, "Connect")
}

func (_mr *_MockConnectionStreamRecorder) Connect() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Connect")
}

func (_m *MockConnectionStream) GetConfig() *config.Config {
	ret := _m.ctrl.Call(_m, "GetConfig")
	ret0, _ := ret[0].(*config.Config)
	return ret0
}

func (_mr *_MockConnectionStreamRecorder) GetConfig() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetConfig")
}

func (_m *MockConnectionStream) GetConnections() map[string]Connection {
	ret := _m.ctrl.Call(_m, "GetConnections")
	ret0, _ := ret[0].(map[string]Connection)
	return ret0
}

func (_mr *_MockConnectionStreamRecorder) GetConnections() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetConnections")
}

func (_m *MockConnectionStream) GetContext() context.Context {
	ret := _m.ctrl.Call(_m, "GetContext")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

func (_mr *_MockConnectionStreamRecorder) GetContext() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetContext")
}

func (_m *MockConnectionStream) GetSchedulers() map[string]Scheduler {
	ret := _m.ctrl.Call(_m, "GetSchedulers")
	ret0, _ := ret[0].(map[string]Scheduler)
	return ret0
}

func (_mr *_MockConnectionStreamRecorder) GetSchedulers() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetSchedulers")
}

func (_m *MockConnectionStream) RegisterConnection(_param0 string, _param1 Connection) error {
	ret := _m.ctrl.Call(_m, "RegisterConnection", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockConnectionStreamRecorder) RegisterConnection(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "RegisterConnection", arg0, arg1)
}

func (_m *MockConnectionStream) SendMetrics(_param0 *check.CheckResultSet) error {
	ret := _m.ctrl.Call(_m, "SendMetrics", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockConnectionStreamRecorder) SendMetrics(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SendMetrics", arg0)
}

func (_m *MockConnectionStream) Stop() {
	_m.ctrl.Call(_m, "Stop")
}

func (_mr *_MockConnectionStreamRecorder) Stop() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Stop")
}

func (_m *MockConnectionStream) StopNotify() chan struct{} {
	ret := _m.ctrl.Call(_m, "StopNotify")
	ret0, _ := ret[0].(chan struct{})
	return ret0
}

func (_mr *_MockConnectionStreamRecorder) StopNotify() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "StopNotify")
}

func (_m *MockConnectionStream) WaitCh() <-chan struct{} {
	ret := _m.ctrl.Call(_m, "WaitCh")
	ret0, _ := ret[0].(<-chan struct{})
	return ret0
}

func (_mr *_MockConnectionStreamRecorder) WaitCh() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "WaitCh")
}