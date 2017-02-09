// Automatically generated by MockGen. DO NOT EDIT!
// Source: net (interfaces: Conn)

package mock_golang

import (
	gomock "github.com/golang/mock/gomock"
	net "net"
	time "time"
)

// Mock of Conn interface
type MockConn struct {
	ctrl     *gomock.Controller
	recorder *_MockConnRecorder
}

// Recorder for MockConn (not exported)
type _MockConnRecorder struct {
	mock *MockConn
}

func NewMockConn(ctrl *gomock.Controller) *MockConn {
	mock := &MockConn{ctrl: ctrl}
	mock.recorder = &_MockConnRecorder{mock}
	return mock
}

func (_m *MockConn) EXPECT() *_MockConnRecorder {
	return _m.recorder
}

func (_m *MockConn) Close() error {
	ret := _m.ctrl.Call(_m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockConnRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Close")
}

func (_m *MockConn) LocalAddr() net.Addr {
	ret := _m.ctrl.Call(_m, "LocalAddr")
	ret0, _ := ret[0].(net.Addr)
	return ret0
}

func (_mr *_MockConnRecorder) LocalAddr() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "LocalAddr")
}

func (_m *MockConn) Read(_param0 []byte) (int, error) {
	ret := _m.ctrl.Call(_m, "Read", _param0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockConnRecorder) Read(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Read", arg0)
}

func (_m *MockConn) RemoteAddr() net.Addr {
	ret := _m.ctrl.Call(_m, "RemoteAddr")
	ret0, _ := ret[0].(net.Addr)
	return ret0
}

func (_mr *_MockConnRecorder) RemoteAddr() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "RemoteAddr")
}

func (_m *MockConn) SetDeadline(_param0 time.Time) error {
	ret := _m.ctrl.Call(_m, "SetDeadline", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockConnRecorder) SetDeadline(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetDeadline", arg0)
}

func (_m *MockConn) SetReadDeadline(_param0 time.Time) error {
	ret := _m.ctrl.Call(_m, "SetReadDeadline", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockConnRecorder) SetReadDeadline(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetReadDeadline", arg0)
}

func (_m *MockConn) SetWriteDeadline(_param0 time.Time) error {
	ret := _m.ctrl.Call(_m, "SetWriteDeadline", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockConnRecorder) SetWriteDeadline(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetWriteDeadline", arg0)
}

func (_m *MockConn) Write(_param0 []byte) (int, error) {
	ret := _m.ctrl.Call(_m, "Write", _param0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockConnRecorder) Write(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Write", arg0)
}
