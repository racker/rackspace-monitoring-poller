//
// Copyright 2016 Rackspace
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS-IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package protocol

import (
	"encoding/json"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/protocol/check"
	"github.com/racker/rackspace-monitoring-poller/utils"
)

///////////////////////////////////////////////////////////////////////////////
// Handshake

type HandshakeParameters struct {
	Token          string              `json:"token"`
	AgentId        string              `json:"agent_id"`
	AgentName      string              `json:"agent_name"`
	ProcessVersion string              `json:"process_version"`
	BundleVersion  string              `json:"bundle_version"`
	ZoneIds        []string            `json:"zone_ids"`
	Features       []map[string]string `json:"features"`
}

type HandshakeRequest struct {
	FrameMsg
	Params HandshakeParameters `json:"params"`
}

func NewHandshakeRequest(cfg *config.Config) Frame {
	f := &HandshakeRequest{}
	f.Version = Version
	f.Method = MethodHandshakeHello
	f.Params.Token = cfg.Token
	f.Params.AgentId = cfg.AgentId
	f.Params.AgentName = cfg.AgentName
	f.Params.ProcessVersion = cfg.ProcessVersion
	f.Params.BundleVersion = cfg.BundleVersion
	f.Params.Features = cfg.Features
	f.Params.ZoneIds = cfg.ZoneIds
	return f
}

func (r HandshakeRequest) Encode() ([]byte, error) {
	return json.Marshal(r)
}

type HandshakeResult struct {
	// HeartbeatInterval indicates to the poller how often it should send a heartbeat.
	// Unit is millisecond.
	HeartbeatInterval uint64 `json:"heartbeat_interval"`
	EntityId          string `json:"entity_id"`
	Channel           string `json:"channel"`
}

type HandshakeResponse struct {
	FrameMsg
	Result HandshakeResult `json:"result"`
}

func DecodeHandshakeResponse(frame *FrameMsg) *HandshakeResponse {
	resp := &HandshakeResponse{}
	resp.SetFromFrameMsg(frame)
	if frame.GetRawResult() != nil {
		json.Unmarshal(frame.GetRawResult(), &resp.Result)
	}
	return resp
}

///////////////////////////////////////////////////////////////////////////////
// Heartbeat

type HeartbeatParameters struct {
	Timestamp int64 `json:"timestamp"`
}

type HeartbeatResult struct {
	Timestamp int64 `json:"timestamp"`
}

type HeartbeatRequest struct {
	FrameMsg
	Params HeartbeatParameters `json:"params"`
}

type HeartbeatResponse struct {
	FrameMsg
	Result HeartbeatResult `json:"result"`
}

func DecodeHeartbeatResponse(frame *FrameMsg) *HeartbeatResponse {
	resp := &HeartbeatResponse{}
	resp.SetFromFrameMsg(frame)
	if frame.GetRawResult() != nil {
		json.Unmarshal(frame.GetRawResult(), &resp.Result)
	}
	return resp
}

func NewHeartbeatRequest() *HeartbeatRequest {
	f := &HeartbeatRequest{}
	f.Version = Version
	f.Method = MethodHeartbeatPost
	f.Params.Timestamp = utils.NowTimestampMillis()
	return f
}

func (r HeartbeatRequest) Encode() ([]byte, error) {
	return json.Marshal(r)
}

func (r HeartbeatRequest) Stringer() string {
	bytes, err := r.Encode()
	if err != nil {
		return ""
	}
	return string(bytes)
}

///////////////////////////////////////////////////////////////////////////////
// Check Schedule Get

type CheckScheduleGet struct {
	FrameMsg
	Params map[string]uint64 `json:"params"`
}

func NewCheckScheduleGet() Frame {
	f := &CheckScheduleGet{}
	f.Version = Version
	f.Method = MethodCheckScheduleGet
	f.Params = map[string]uint64{"blah": 1}
	return f
}

func (r CheckScheduleGet) Encode() ([]byte, error) {
	return json.Marshal(r)
}

///////////////////////////////////////////////////////////////////////////////
// HostInfo

type HostInfoResponse struct {
	FrameMsgCommon

	Result interface{} `json:"result"`
}

func (r *HostInfoResponse) Encode() ([]byte, error) {
	return json.Marshal(r)
}

///////////////////////////////////////////////////////////////////////////////
// Metrics Post

type MetricWrap []map[string]*MetricTVU
type MetricWrapper []MetricWrap

type MetricTVU struct {
	Type  string `json:"t"`
	Value string `json:"v"`
	Unit  string `json:"u"`
}

type MetricsPostRequestParams struct {
	EntityId       string       `json:"entity_id"`
	CheckId        string       `json:"check_id"`
	CheckType      string       `json:"check_type"`
	Metrics        []MetricWrap `json:"metrics"`
	MinCheckPeriod uint64       `json:"min_check_period"`
	State          string       `json:"state"`
	Status         string       `json:"status"`
	Timestamp      int64        `json:"timestamp"`
}

type MetricsPostRequest struct {
	FrameMsg
	Params MetricsPostRequestParams `json:"params"`
}

func (r MetricsPostRequest) Encode() ([]byte, error) {
	return json.Marshal(r)
}

///////////////////////////////////////////////////////////////////////////////
// Poller Prepare (start, block, end)

const (
	PrepareActionStart    = "start"
	PrepareActionRestart  = "restart"
	PrepareActionContinue = "continue"

	PrepareDirectivePrepare = "prepare"
	PrepareDirectiveAbort   = "abort"

	PrepareResultStatusPrepared  = "prepared"
	PrepareResultStatusAborted   = "aborted"
	PrepareResultStatusFailed    = "failed"
	PrepareResultStatusIgnored   = "ignored"
	PrepareResultStatusCommitted = "committed"
)

type PollerPrepareManifest struct {
	// Action is one of PrepareAction* constants
	Action    string `json:"action"`
	Id        string `json:"check_id"`
	CheckType string `json:"type"`
	EntityId  string `json:"entity_id"`
	ZoneId    string `json:"zone_id"`
}

// PollerPrepareRequest abstracts common info across the "poller.prepare.*" methods
type PollerPrepareRequest interface {
	GetPreparationZoneId() string
	GetPreparationVersion() int
}

// PollerPrepareStartParams is the params of a message with method=MethodPollerPrepare
type PollerPrepareStartParams struct {
	ZoneId   string                  `json:"zone_id"`
	Version  int                     `json:"version"`
	Manifest []PollerPrepareManifest `json:"manifest"`
}

type PollerPrepareStartRequest struct {
	FrameMsg
	Params PollerPrepareStartParams `json:"params"`
}

func (req *PollerPrepareStartRequest) String() string {
	json, _ := json.Marshal(req)
	return string(json)
}

func (req *PollerPrepareStartRequest) GetPreparationVersion() int {
	return req.Params.Version
}
func (req *PollerPrepareStartRequest) GetPreparationZoneId() string {
	return req.Params.ZoneId
}

func DecodePollerPrepareStartRequest(frame *FrameMsg) *PollerPrepareStartRequest {
	req := &PollerPrepareStartRequest{}
	req.SetFromFrameMsg(frame)
	if frame.GetRawParams() != nil {
		json.Unmarshal(frame.GetRawParams(), &req.Params)
	}
	return req
}

// PollerPrepareBlockParams is the params of a message with method=MethodPollerPrepareBlock
type PollerPrepareBlockParams struct {
	ZoneId  string           `json:"zone_id"`
	Version int              `json:"version"`
	Block   []*check.CheckIn `json:"block"`
}

type PollerPrepareBlockRequest struct {
	FrameMsg
	Params PollerPrepareBlockParams `json:"params"`
}

func (req *PollerPrepareBlockRequest) String() string {
	json, _ := json.Marshal(req)
	return string(json)
}

func (req *PollerPrepareBlockRequest) GetPreparationVersion() int {
	return req.Params.Version
}
func (req *PollerPrepareBlockRequest) GetPreparationZoneId() string {
	return req.Params.ZoneId
}

func DecodePollerPrepareBlockRequest(frame *FrameMsg) *PollerPrepareBlockRequest {
	req := &PollerPrepareBlockRequest{}
	req.SetFromFrameMsg(frame)
	if frame.GetRawParams() != nil {
		json.Unmarshal(frame.GetRawParams(), &req.Params)
	}
	return req
}

// PollerPrepareBlockParams is the params of a message with method=MethodPollerPrepareEnd
type PollerPrepareEndParams struct {
	ZoneId  string `json:"zone_id"`
	Version int    `json:"version"`
	// Directive is one of PrepareDirective* constants
	Directive string `json:"directive"`
}

type PollerPrepareEndRequest struct {
	FrameMsg
	Params PollerPrepareEndParams `json:"params"`
}

func (req *PollerPrepareEndRequest) String() string {
	json, _ := json.Marshal(req)
	return string(json)
}

func (req *PollerPrepareEndRequest) GetPreparationVersion() int {
	return req.Params.Version
}
func (req *PollerPrepareEndRequest) GetPreparationZoneId() string {
	return req.Params.ZoneId
}

func DecodePollerPrepareEndRequest(frame *FrameMsg) *PollerPrepareEndRequest {
	req := &PollerPrepareEndRequest{}
	req.SetFromFrameMsg(frame)
	if frame.GetRawParams() != nil {
		json.Unmarshal(frame.GetRawParams(), &req.Params)
	}
	return req
}

// PollerCommitParams is the params of a message with method=MethodPollerCommit
type PollerCommitParams struct {
	ZoneId  string `json:"zone_id"`
	Version int    `json:"version"`
}

type PollerCommitRequest struct {
	FrameMsg
	Params PollerCommitParams `json:"params"`
}

func DecodePollerCommitRequest(frame *FrameMsg) *PollerCommitRequest {
	req := &PollerCommitRequest{}
	req.SetFromFrameMsg(frame)
	if frame.GetRawParams() != nil {
		json.Unmarshal(frame.GetRawParams(), &req.Params)
	}
	return req
}

type PollerPrepareResult struct {
	ZoneId  string `json:"zone_id"`
	Version int    `json:"version"`
	// Status is one of PrepareResultStatus* constants
	Status  string `json:"status"`
	Details string `json:"details"`
}

type PollerPrepareResponse struct {
	FrameMsg
	Result PollerPrepareResult `json:"result"`
}

func NewPollerPrepareResponse(source *FrameMsg, result PollerPrepareResult) Frame {
	resp := &FrameMsg{}
	resp.SetResponseFrameMsg(source)

	raw, err := json.Marshal(result)
	if err != nil {
		return nil
	}

	resp.RawResult = json.RawMessage(raw)

	return resp
}

type PollerCommitResult struct {
	ZoneId  string `json:"zone_id"`
	Version int    `json:"version"`
	// Status is one of PrepareResultStatus* constants
	Status  string `json:"status"`
	Details string `json:"details"`
}

func NewPollerPrepareCommitResponse(source *FrameMsg, result PollerCommitResult) Frame {
	resp := &FrameMsg{}
	resp.SetResponseFrameMsg(source)

	raw, err := json.Marshal(result)
	if err != nil {
		return nil
	}

	resp.RawResult = json.RawMessage(raw)

	return resp
}

type PollerCommitResponse struct {
	FrameMsg
	Result PollerCommitResult `json:"result"`
}
