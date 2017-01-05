// Automatically generated by MockGen. DO NOT EDIT!
// Source: poller/connection.go

package poller

import (
	context "context"
	"crypto/tls"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// Mock of ConnectionInterface interface
type MockConnectionInterface struct {
	ctrl     *gomock.Controller
	recorder *_MockConnectionInterfaceRecorder
}

// Recorder for MockConnectionInterface (not exported)
type _MockConnectionInterfaceRecorder struct {
	mock *MockConnectionInterface
}

func NewMockConnectionInterface(ctrl *gomock.Controller) *MockConnectionInterface {
	mock := &MockConnectionInterface{ctrl: ctrl}
	mock.recorder = &_MockConnectionInterfaceRecorder{mock}
	return mock
}

func (_m *MockConnectionInterface) EXPECT() *_MockConnectionInterfaceRecorder {
	return _m.recorder
}

func (_m *MockConnectionInterface) GetStream() ConnectionStreamInterface {
	ret := _m.ctrl.Call(_m, "GetStream")
	ret0, _ := ret[0].(ConnectionStreamInterface)
	return ret0
}

func (_mr *_MockConnectionInterfaceRecorder) GetStream() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetStream")
}

func (_m *MockConnectionInterface) GetSession() SessionInterface {
	ret := _m.ctrl.Call(_m, "GetSession")
	ret0, _ := ret[0].(SessionInterface)
	return ret0
}

func (_mr *_MockConnectionInterfaceRecorder) GetSession() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetSession")
}

func (_m *MockConnectionInterface) SetReadDeadline(deadline time.Time) {
	_m.ctrl.Call(_m, "SetReadDeadline", deadline)
}

func (_mr *_MockConnectionInterfaceRecorder) SetReadDeadline(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetReadDeadline", arg0)
}

func (_m *MockConnectionInterface) SetWriteDeadline(deadline time.Time) {
	_m.ctrl.Call(_m, "SetWriteDeadline", deadline)
}

func (_mr *_MockConnectionInterfaceRecorder) SetWriteDeadline(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetWriteDeadline", arg0)
}

func (_m *MockConnectionInterface) Connect(ctx context.Context, tlsConfig *tls.Config) error {
	ret := _m.ctrl.Call(_m, "Connect", ctx, tlsConfig)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockConnectionInterfaceRecorder) Connect(arg0 interface{}, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Connect", arg0, arg1)
}

func (_m *MockConnectionInterface) Close() {
	_m.ctrl.Call(_m, "Close")
}

func (_mr *_MockConnectionInterfaceRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Close")
}

func (_m *MockConnectionInterface) Wait() {
	_m.ctrl.Call(_m, "Wait")
}

func (_mr *_MockConnectionInterfaceRecorder) Wait() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Wait")
}

func (_m *MockConnectionInterface) GetConnection() *Connection {
	ret := _m.ctrl.Call(_m, "GetConnection")
	ret0, _ := ret[0].(*Connection)
	return ret0
}

func (_mr *_MockConnectionInterfaceRecorder) GetConnection() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetConnection")
}