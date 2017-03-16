// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/racker/rackspace-monitoring-poller/poller (interfaces: LogPrefixGetter,ConnectionStream,Connection,Session,CheckScheduler,CheckExecutor,Scheduler,ChecksReconciler)

package poller_test

import (
	context "context"
	tls "crypto/tls"
	gomock "github.com/golang/mock/gomock"
	check "github.com/racker/rackspace-monitoring-poller/check"
	config "github.com/racker/rackspace-monitoring-poller/config"
	poller "github.com/racker/rackspace-monitoring-poller/poller"
	protocol "github.com/racker/rackspace-monitoring-poller/protocol"
	utils "github.com/racker/rackspace-monitoring-poller/utils"
	io "io"
	time "time"
)

// Mock of LogPrefixGetter interface
type MockLogPrefixGetter struct {
	ctrl     *gomock.Controller
	recorder *_MockLogPrefixGetterRecorder
}

// Recorder for MockLogPrefixGetter (not exported)
type _MockLogPrefixGetterRecorder struct {
	mock *MockLogPrefixGetter
}

func NewMockLogPrefixGetter(ctrl *gomock.Controller) *MockLogPrefixGetter {
	mock := &MockLogPrefixGetter{ctrl: ctrl}
	mock.recorder = &_MockLogPrefixGetterRecorder{mock}
	return mock
}

func (_m *MockLogPrefixGetter) EXPECT() *_MockLogPrefixGetterRecorder {
	return _m.recorder
}

func (_m *MockLogPrefixGetter) GetLogPrefix() string {
	ret := _m.ctrl.Call(_m, "GetLogPrefix")
	ret0, _ := ret[0].(string)
	return ret0
}

func (_mr *_MockLogPrefixGetterRecorder) GetLogPrefix() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetLogPrefix")
}

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

func (_m *MockConnectionStream) DeregisterEventConsumer(_param0 utils.EventConsumer) {
	_m.ctrl.Call(_m, "DeregisterEventConsumer", _param0)
}

func (_mr *_MockConnectionStreamRecorder) DeregisterEventConsumer(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DeregisterEventConsumer", arg0)
}

func (_m *MockConnectionStream) Done() <-chan struct{} {
	ret := _m.ctrl.Call(_m, "Done")
	ret0, _ := ret[0].(<-chan struct{})
	return ret0
}

func (_mr *_MockConnectionStreamRecorder) Done() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Done")
}

func (_m *MockConnectionStream) RegisterEventConsumer(_param0 utils.EventConsumer) {
	_m.ctrl.Call(_m, "RegisterEventConsumer", _param0)
}

func (_mr *_MockConnectionStreamRecorder) RegisterEventConsumer(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "RegisterEventConsumer", arg0)
}

func (_m *MockConnectionStream) SendMetrics(_param0 *check.ResultSet) {
	_m.ctrl.Call(_m, "SendMetrics", _param0)
}

func (_mr *_MockConnectionStreamRecorder) SendMetrics(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SendMetrics", arg0)
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

func (_m *MockConnection) Close() {
	_m.ctrl.Call(_m, "Close")
}

func (_mr *_MockConnectionRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Close")
}

func (_m *MockConnection) Connect(_param0 context.Context, _param1 *config.Config, _param2 *tls.Config) error {
	ret := _m.ctrl.Call(_m, "Connect", _param0, _param1, _param2)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockConnectionRecorder) Connect(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Connect", arg0, arg1, arg2)
}

func (_m *MockConnection) Done() <-chan struct{} {
	ret := _m.ctrl.Call(_m, "Done")
	ret0, _ := ret[0].(<-chan struct{})
	return ret0
}

func (_mr *_MockConnectionRecorder) Done() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Done")
}

func (_m *MockConnection) GetClockOffset() int64 {
	ret := _m.ctrl.Call(_m, "GetClockOffset")
	ret0, _ := ret[0].(int64)
	return ret0
}

func (_mr *_MockConnectionRecorder) GetClockOffset() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetClockOffset")
}

func (_m *MockConnection) GetFarendReader() io.Reader {
	ret := _m.ctrl.Call(_m, "GetFarendReader")
	ret0, _ := ret[0].(io.Reader)
	return ret0
}

func (_mr *_MockConnectionRecorder) GetFarendReader() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetFarendReader")
}

func (_m *MockConnection) GetFarendWriter() io.Writer {
	ret := _m.ctrl.Call(_m, "GetFarendWriter")
	ret0, _ := ret[0].(io.Writer)
	return ret0
}

func (_mr *_MockConnectionRecorder) GetFarendWriter() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetFarendWriter")
}

func (_m *MockConnection) GetGUID() string {
	ret := _m.ctrl.Call(_m, "GetGUID")
	ret0, _ := ret[0].(string)
	return ret0
}

func (_mr *_MockConnectionRecorder) GetGUID() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetGUID")
}

func (_m *MockConnection) GetLatency() int64 {
	ret := _m.ctrl.Call(_m, "GetLatency")
	ret0, _ := ret[0].(int64)
	return ret0
}

func (_mr *_MockConnectionRecorder) GetLatency() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetLatency")
}

func (_m *MockConnection) GetLogPrefix() string {
	ret := _m.ctrl.Call(_m, "GetLogPrefix")
	ret0, _ := ret[0].(string)
	return ret0
}

func (_mr *_MockConnectionRecorder) GetLogPrefix() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetLogPrefix")
}

func (_m *MockConnection) GetSession() poller.Session {
	ret := _m.ctrl.Call(_m, "GetSession")
	ret0, _ := ret[0].(poller.Session)
	return ret0
}

func (_mr *_MockConnectionRecorder) GetSession() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetSession")
}

func (_m *MockConnection) SetReadDeadline(_param0 time.Time) {
	_m.ctrl.Call(_m, "SetReadDeadline", _param0)
}

func (_mr *_MockConnectionRecorder) SetReadDeadline(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetReadDeadline", arg0)
}

func (_m *MockConnection) SetWriteDeadline(_param0 time.Time) {
	_m.ctrl.Call(_m, "SetWriteDeadline", _param0)
}

func (_mr *_MockConnectionRecorder) SetWriteDeadline(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetWriteDeadline", arg0)
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

func (_m *MockSession) Close() {
	_m.ctrl.Call(_m, "Close")
}

func (_mr *_MockSessionRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Close")
}

func (_m *MockSession) Done() <-chan struct{} {
	ret := _m.ctrl.Call(_m, "Done")
	ret0, _ := ret[0].(<-chan struct{})
	return ret0
}

func (_mr *_MockSessionRecorder) Done() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Done")
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

func (_m *MockSession) Respond(_param0 protocol.Frame) {
	_m.ctrl.Call(_m, "Respond", _param0)
}

func (_mr *_MockSessionRecorder) Respond(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Respond", arg0)
}

func (_m *MockSession) Send(_param0 protocol.Frame) {
	_m.ctrl.Call(_m, "Send", _param0)
}

func (_mr *_MockSessionRecorder) Send(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Send", arg0)
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

func (_m *MockCheckScheduler) Schedule(_param0 check.Check) {
	_m.ctrl.Call(_m, "Schedule", _param0)
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

func (_m *MockCheckExecutor) Execute(_param0 check.Check) {
	_m.ctrl.Call(_m, "Execute", _param0)
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

func (_m *MockScheduler) Close() {
	_m.ctrl.Call(_m, "Close")
}

func (_mr *_MockSchedulerRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Close")
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

func (_m *MockScheduler) GetScheduledChecks() []check.Check {
	ret := _m.ctrl.Call(_m, "GetScheduledChecks")
	ret0, _ := ret[0].([]check.Check)
	return ret0
}

func (_mr *_MockSchedulerRecorder) GetScheduledChecks() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetScheduledChecks")
}

func (_m *MockScheduler) GetZoneID() string {
	ret := _m.ctrl.Call(_m, "GetZoneID")
	ret0, _ := ret[0].(string)
	return ret0
}

func (_mr *_MockSchedulerRecorder) GetZoneID() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetZoneID")
}

func (_m *MockScheduler) ReconcileChecks(_param0 poller.ChecksPrepared) {
	_m.ctrl.Call(_m, "ReconcileChecks", _param0)
}

func (_mr *_MockSchedulerRecorder) ReconcileChecks(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ReconcileChecks", arg0)
}

func (_m *MockScheduler) SendMetrics(_param0 *check.ResultSet) {
	_m.ctrl.Call(_m, "SendMetrics", _param0)
}

func (_mr *_MockSchedulerRecorder) SendMetrics(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SendMetrics", arg0)
}

func (_m *MockScheduler) ValidateChecks(_param0 poller.ChecksPreparing) error {
	ret := _m.ctrl.Call(_m, "ValidateChecks", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockSchedulerRecorder) ValidateChecks(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ValidateChecks", arg0)
}

// Mock of ChecksReconciler interface
type MockChecksReconciler struct {
	ctrl     *gomock.Controller
	recorder *_MockChecksReconcilerRecorder
}

// Recorder for MockChecksReconciler (not exported)
type _MockChecksReconcilerRecorder struct {
	mock *MockChecksReconciler
}

func NewMockChecksReconciler(ctrl *gomock.Controller) *MockChecksReconciler {
	mock := &MockChecksReconciler{ctrl: ctrl}
	mock.recorder = &_MockChecksReconcilerRecorder{mock}
	return mock
}

func (_m *MockChecksReconciler) EXPECT() *_MockChecksReconcilerRecorder {
	return _m.recorder
}

func (_m *MockChecksReconciler) ReconcileChecks(_param0 poller.ChecksPrepared) {
	_m.ctrl.Call(_m, "ReconcileChecks", _param0)
}

func (_mr *_MockChecksReconcilerRecorder) ReconcileChecks(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ReconcileChecks", arg0)
}

func (_m *MockChecksReconciler) ValidateChecks(_param0 poller.ChecksPreparing) error {
	ret := _m.ctrl.Call(_m, "ValidateChecks", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockChecksReconcilerRecorder) ValidateChecks(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ValidateChecks", arg0)
}
