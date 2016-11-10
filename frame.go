package main

import (
	"encoding/json"
	"sync/atomic"
)

type Error struct {
	Code    uint64 `json:"code"`
	Message string `json:"message"`
}

type Frame interface {
	Encode() ([]byte, error)
	GetVersion() string
	SetVersion(string)
	GetId() uint64
	SetId(session *Session)
	SetRawId(uint64)
	GetTarget() string
	SetTarget(string)
	GetSource() string
	SetSource(string)
	GetMethod() string
	SetMethod(string)
	GetRawParams() *json.RawMessage
	GetRawResult() *json.RawMessage
	GetError() *Error
	SetFromFrameMsg(*FrameMsg)
}

type FrameMsgCommon struct {
	Version   string           `json:"v"`
	Id        uint64           `json:"id"`
	Target    string           `json:"target"`
	Source    string           `json:"source"`
	Method    string           `json:"method,omitempty"`

	Error     *Error           `json:"error,omitempty"`
}

type FrameMsg struct {
	FrameMsgCommon
	RawParams *json.RawMessage `json:"params,omitempty"`
	RawResult *json.RawMessage `json:"result,omitempty"`
}

func (f *FrameMsgCommon) SetRawId(id uint64) {
	f.Id = id
}

func (f *FrameMsgCommon) SetResponseFrameMsg(source *FrameMsg) {
	f.Id = source.Id
	f.Source = source.Source
	f.Target = source.Target
	f.Version = source.Version
}

func (f *FrameMsgCommon) SetFromFrameMsg(source *FrameMsg) {
	f.Id = source.Id
	f.Method = source.Method
	f.Source = source.Source
	f.Target = source.Target
	f.Version = source.Version
}

func (f *FrameMsg) SetFromFrameMsg(source *FrameMsg) {
	f.Id = source.Id
	f.Method = source.Method
	f.Source = source.Source
	f.Target = source.Target
	f.Version = source.Version
	f.RawResult = source.RawResult
	f.RawParams = source.RawParams
}

func (r *FrameMsgCommon) Encode() ([]byte, error) {
	return json.Marshal(r)
}

func (f *FrameMsgCommon) GetError() *Error {
	return f.Error
}

func (f *FrameMsgCommon) SetSource(source string) {
	f.Source = source
}

func (f *FrameMsgCommon) GetId() uint64 {
	return f.Id
}

func (f *FrameMsgCommon) GetMethod() string {
	return f.Method
}

func (f *FrameMsgCommon) GetSource() string {
	return f.Source
}

func (f *FrameMsgCommon) GetTarget() string {
	return f.Target
}

func (f *FrameMsgCommon) GetVersion() string {
	return f.Version
}

func (f *FrameMsgCommon) SetMethod(method string) {
	f.Method = method
}

func (f *FrameMsgCommon) SetVersion(version string) {
	f.Version = version
}

func (f *FrameMsgCommon) GetRawParams() *json.RawMessage {
	return nil
}

func (f *FrameMsg) GetRawParams() *json.RawMessage {
	return f.RawParams
}

func (f *FrameMsgCommon) GetRawResult() *json.RawMessage {
	return nil
}

func (f *FrameMsg) GetRawResult() *json.RawMessage {
	return f.RawResult
}

func (f *FrameMsgCommon) SetId(s *Session) {
get_id:
	f.Id = atomic.LoadUint64(&s.seq)
	if !atomic.CompareAndSwapUint64(&s.seq, f.Id, f.Id+1) {
		goto get_id
	}
}

func (f *FrameMsgCommon) SetTarget(target string) {
	f.Target = target
}
