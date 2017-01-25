// Automatically generated by MockGen. DO NOT EDIT!
// Source: poller/poller.go

package poller

import (
	"context"
	"crypto/tls"
	"github.com/golang/mock/gomock"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"io"
	"time"
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

func (_m *MockConnectionStream) GetConfig() *config.Config {
	ret := _m.ctrl.Call(_m, "GetConfig")
	ret0, _ := ret[0].(*config.Config)
	return ret0
}

func (_mr *_MockConnectionStreamRecorder) GetConfig() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetConfig")
}

func (_m *MockConnectionStream) RegisterConnection(qry string, conn Connection) error {
	ret := _m.ctrl.Call(_m, "RegisterConnection", qry, conn)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockConnectionStreamRecorder) RegisterConnection(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "RegisterConnection", arg0, arg1)
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

func (_m *MockConnectionStream) GetSchedulers() map[string]Scheduler {
	ret := _m.ctrl.Call(_m, "GetSchedulers")
	ret0, _ := ret[0].(map[string]Scheduler)
	return ret0
}

func (_mr *_MockConnectionStreamRecorder) GetSchedulers() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetSchedulers")
}

func (_m *MockConnectionStream) GetContext() context.Context {
	ret := _m.ctrl.Call(_m, "GetContext")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

func (_mr *_MockConnectionStreamRecorder) GetContext() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetContext")
}

func (_m *MockConnectionStream) SendMetrics(crs *check.ResultSet) error {
	ret := _m.ctrl.Call(_m, "SendMetrics", crs)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockConnectionStreamRecorder) SendMetrics(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SendMetrics", arg0)
}

func (_m *MockConnectionStream) Connect() {
	_m.ctrl.Call(_m, "Connect")
}

func (_mr *_MockConnectionStreamRecorder) Connect() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Connect")
}

func (_m *MockConnectionStream) WaitCh() <-chan struct{} {
	ret := _m.ctrl.Call(_m, "WaitCh")
	ret0, _ := ret[0].(<-chan struct{})
	return ret0
}

func (_mr *_MockConnectionStreamRecorder) WaitCh() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "WaitCh")
}

func (_m *MockConnectionStream) GetConnections() map[string]Connection {
	ret := _m.ctrl.Call(_m, "GetConnections")
	ret0, _ := ret[0].(map[string]Connection)
	return ret0
}

func (_mr *_MockConnectionStreamRecorder) GetConnections() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetConnections")
}

// Mock of Connection interface
type MockConnection struct {
	ctrl     *gomock.Controller
	recorder *_MockConnectionRecorder
}

// Recorder for MockConnection (not exported)
type _MockConnectionRecorder struct {
	mock *MockConnection
}

func NewMockConnection(ctrl *gomock.Controller) *MockConnection {
	mock := &MockConnection{ctrl: ctrl}
	mock.recorder = &_MockConnectionRecorder{mock}
	return mock
}

func (_m *MockConnection) EXPECT() *_MockConnectionRecorder {
	return _m.recorder
}

func (_m *MockConnection) GetStream() ConnectionStream {
	ret := _m.ctrl.Call(_m, "GetStream")
	ret0, _ := ret[0].(ConnectionStream)
	return ret0
}

func (_mr *_MockConnectionRecorder) GetStream() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetStream")
}

func (_m *MockConnection) GetSession() Session {
	ret := _m.ctrl.Call(_m, "GetSession")
	ret0, _ := ret[0].(Session)
	return ret0
}

func (_mr *_MockConnectionRecorder) GetSession() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetSession")
}

func (_m *MockConnection) SetReadDeadline(deadline time.Time) {
	_m.ctrl.Call(_m, "SetReadDeadline", deadline)
}

func (_mr *_MockConnectionRecorder) SetReadDeadline(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetReadDeadline", arg0)
}

func (_m *MockConnection) SetWriteDeadline(deadline time.Time) {
	_m.ctrl.Call(_m, "SetWriteDeadline", deadline)
}

func (_mr *_MockConnectionRecorder) SetWriteDeadline(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetWriteDeadline", arg0)
}

func (_m *MockConnection) Connect(ctx context.Context, tlsConfig *tls.Config) error {
	ret := _m.ctrl.Call(_m, "Connect", ctx, tlsConfig)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockConnectionRecorder) Connect(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Connect", arg0, arg1)
}

func (_m *MockConnection) Close() {
	_m.ctrl.Call(_m, "Close")
}

func (_mr *_MockConnectionRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Close")
}

func (_m *MockConnection) Wait() {
	_m.ctrl.Call(_m, "Wait")
}

func (_mr *_MockConnectionRecorder) Wait() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Wait")
}

func (_m *MockConnection) GetConnection() io.ReadWriteCloser {
	ret := _m.ctrl.Call(_m, "GetConnection")
	ret0, _ := ret[0].(io.ReadWriteCloser)
	return ret0
}

func (_mr *_MockConnectionRecorder) GetConnection() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetConnection")
}

func (_m *MockConnection) GetGUID() string {
	ret := _m.ctrl.Call(_m, "GetGUID")
	ret0, _ := ret[0].(string)
	return ret0
}

func (_mr *_MockConnectionRecorder) GetGUID() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetGUID")
}

// Mock of Session interface
type MockSession struct {
	ctrl     *gomock.Controller
	recorder *_MockSessionRecorder
}

// Recorder for MockSession (not exported)
type _MockSessionRecorder struct {
	mock *MockSession
}

func NewMockSession(ctrl *gomock.Controller) *MockSession {
	mock := &MockSession{ctrl: ctrl}
	mock.recorder = &_MockSessionRecorder{mock}
	return mock
}

func (_m *MockSession) EXPECT() *_MockSessionRecorder {
	return _m.recorder
}

func (_m *MockSession) Auth() {
	_m.ctrl.Call(_m, "Auth")
}

func (_mr *_MockSessionRecorder) Auth() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Auth")
}

func (_m *MockSession) Send(msg protocol.Frame) {
	_m.ctrl.Call(_m, "Send", msg)
}

func (_mr *_MockSessionRecorder) Send(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Send", arg0)
}

func (_m *MockSession) Respond(msg protocol.Frame) {
	_m.ctrl.Call(_m, "Respond", msg)
}

func (_mr *_MockSessionRecorder) Respond(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Respond", arg0)
}

func (_m *MockSession) SetHeartbeatInterval(timeout uint64) {
	_m.ctrl.Call(_m, "SetHeartbeatInterval", timeout)
}

func (_mr *_MockSessionRecorder) SetHeartbeatInterval(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetHeartbeatInterval", arg0)
}

func (_m *MockSession) GetReadDeadline() time.Time {
	ret := _m.ctrl.Call(_m, "GetReadDeadline")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

func (_mr *_MockSessionRecorder) GetReadDeadline() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetReadDeadline")
}

func (_m *MockSession) GetWriteDeadline() time.Time {
	ret := _m.ctrl.Call(_m, "GetWriteDeadline")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

func (_mr *_MockSessionRecorder) GetWriteDeadline() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetWriteDeadline")
}

func (_m *MockSession) Close() {
	_m.ctrl.Call(_m, "Close")
}

func (_mr *_MockSessionRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Close")
}

func (_m *MockSession) Wait() {
	_m.ctrl.Call(_m, "Wait")
}

func (_mr *_MockSessionRecorder) Wait() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Wait")
}

func (_m *MockSession) GetClockOffset() int64 {
	ret := _m.ctrl.Call(_m, "GetClockOffset")
	ret0, _ := ret[0].(int64)
	return ret0
}

func (_mr *_MockSessionRecorder) GetClockOffset() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetClockOffset")
}

func (_m *MockSession) GetLatency() int64 {
	ret := _m.ctrl.Call(_m, "GetLatency")
	ret0, _ := ret[0].(int64)
	return ret0
}

func (_mr *_MockSessionRecorder) GetLatency() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetLatency")
}

// Mock of CheckScheduler interface
type MockCheckScheduler struct {
	ctrl     *gomock.Controller
	recorder *_MockCheckSchedulerRecorder
}

// Recorder for MockCheckScheduler (not exported)
type _MockCheckSchedulerRecorder struct {
	mock *MockCheckScheduler
}

func NewMockCheckScheduler(ctrl *gomock.Controller) *MockCheckScheduler {
	mock := &MockCheckScheduler{ctrl: ctrl}
	mock.recorder = &_MockCheckSchedulerRecorder{mock}
	return mock
}

func (_m *MockCheckScheduler) EXPECT() *_MockCheckSchedulerRecorder {
	return _m.recorder
}

func (_m *MockCheckScheduler) Schedule(ch check.Check) {
	_m.ctrl.Call(_m, "Schedule", ch)
}

func (_mr *_MockCheckSchedulerRecorder) Schedule(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Schedule", arg0)
}

// Mock of CheckExecutor interface
type MockCheckExecutor struct {
	ctrl     *gomock.Controller
	recorder *_MockCheckExecutorRecorder
}

// Recorder for MockCheckExecutor (not exported)
type _MockCheckExecutorRecorder struct {
	mock *MockCheckExecutor
}

func NewMockCheckExecutor(ctrl *gomock.Controller) *MockCheckExecutor {
	mock := &MockCheckExecutor{ctrl: ctrl}
	mock.recorder = &_MockCheckExecutorRecorder{mock}
	return mock
}

func (_m *MockCheckExecutor) EXPECT() *_MockCheckExecutorRecorder {
	return _m.recorder
}

func (_m *MockCheckExecutor) Execute(ch check.Check) {
	_m.ctrl.Call(_m, "Execute", ch)
}

func (_mr *_MockCheckExecutorRecorder) Execute(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Execute", arg0)
}

// Mock of Scheduler interface
type MockScheduler struct {
	ctrl     *gomock.Controller
	recorder *_MockSchedulerRecorder
}

// Recorder for MockScheduler (not exported)
type _MockSchedulerRecorder struct {
	mock *MockScheduler
}

func NewMockScheduler(ctrl *gomock.Controller) *MockScheduler {
	mock := &MockScheduler{ctrl: ctrl}
	mock.recorder = &_MockSchedulerRecorder{mock}
	return mock
}

func (_m *MockScheduler) EXPECT() *_MockSchedulerRecorder {
	return _m.recorder
}

func (_m *MockScheduler) GetInput() chan protocol.Frame {
	ret := _m.ctrl.Call(_m, "GetInput")
	ret0, _ := ret[0].(chan protocol.Frame)
	return ret0
}

func (_mr *_MockSchedulerRecorder) GetInput() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetInput")
}

func (_m *MockScheduler) Close() {
	_m.ctrl.Call(_m, "Close")
}

func (_mr *_MockSchedulerRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Close")
}

func (_m *MockScheduler) SendMetrics(crs *check.ResultSet) {
	_m.ctrl.Call(_m, "SendMetrics", crs)
}

func (_mr *_MockSchedulerRecorder) SendMetrics(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SendMetrics", arg0)
}

func (_m *MockScheduler) Register(ch check.Check) error {
	ret := _m.ctrl.Call(_m, "Register", ch)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockSchedulerRecorder) Register(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Register", arg0)
}

func (_m *MockScheduler) RunFrameConsumer() {
	_m.ctrl.Call(_m, "RunFrameConsumer")
}

func (_mr *_MockSchedulerRecorder) RunFrameConsumer() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "RunFrameConsumer")
}

func (_m *MockScheduler) GetZoneID() string {
	ret := _m.ctrl.Call(_m, "GetZoneID")
	ret0, _ := ret[0].(string)
	return ret0
}

func (_mr *_MockSchedulerRecorder) GetZoneID() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetZoneID")
}

func (_m *MockScheduler) GetContext() (context.Context, context.CancelFunc) {
	ret := _m.ctrl.Call(_m, "GetContext")
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(context.CancelFunc)
	return ret0, ret1
}

func (_mr *_MockSchedulerRecorder) GetContext() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetContext")
}

func (_m *MockScheduler) GetChecks() map[string]check.Check {
	ret := _m.ctrl.Call(_m, "GetChecks")
	ret0, _ := ret[0].(map[string]check.Check)
	return ret0
}

func (_mr *_MockSchedulerRecorder) GetChecks() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetChecks")
}
