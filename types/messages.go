package types

import (
	"encoding/json"
	"fmt"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/racker/rackspace-monitoring-poller/hostinfo"
)

///////////////////////////////////////////////////////////////////////////////
// Handshake

type HandshakeParameters struct {
	Token          string              `json:"token"`
	AgentId        string              `json:"agent_id"`
	AgentName      string              `json:"agent_name"`
	ProcessVersion string              `json:"process_version"`
	BundleVersion  string              `json:"bundle_version"`
	Features       []map[string]string `json:"features"`
}

type HandshakeRequest struct {
	FrameMsg
	Params HandshakeParameters `json:"params"`
}

func NewHandshakeRequest(cfg *Config) Frame {
	f := &HandshakeRequest{}
	f.Version = "1"
	f.Method = "handshake.hello"
	f.Params.Token = cfg.Token
	f.Params.AgentId = cfg.AgentId
	f.Params.AgentName = cfg.AgentName
	f.Params.ProcessVersion = cfg.ProcessVersion
	f.Params.BundleVersion = cfg.BundleVersion
	f.Params.Features = cfg.Features
	return f
}

func (r HandshakeRequest) Encode() ([]byte, error) {
	return json.Marshal(r)
}

type HandshakeResult struct {
	HandshakeInterval uint64 `json:"heartbeat_interval"`
	EntityId          string `json:"entity_id"`
	Channel           string `json:"channel"`
}

type HandshakeResponse struct {
	FrameMsg
	Result HandshakeResult `json:"result"`
}

func NewHandshakeResponse(frame *FrameMsg) *HandshakeResponse {
	resp := &HandshakeResponse{}
	resp.SetFromFrameMsg(frame)
	if frame.GetRawResult() != nil {
		json.Unmarshal(*frame.GetRawResult(), &resp.Result)
	}
	return resp
}

type HeartbeatParameters struct {
	Timestamp int64 `json:"timestamp"`
}

type HeartbeatRequest struct {
	FrameMsg
	Params HeartbeatParameters `json:"params"`
}

///////////////////////////////////////////////////////////////////////////////
// Heartbeat

func NewHeartbeat() Frame {
	f := &HeartbeatRequest{}
	f.Version = "1"
	f.Method = "heartbeat.post"
	f.Params.Timestamp = utils.NowTimestampMillis()
	return f
}

func (r HeartbeatRequest) Encode() ([]byte, error) {
	return json.Marshal(r)
}

type CheckScheduleGet struct {
	FrameMsg
	Params map[string]uint64 `json:"params"`
}

///////////////////////////////////////////////////////////////////////////////
// Check Schedule Get

func NewCheckScheduleGet() Frame {
	f := &CheckScheduleGet{}
	f.Version = "1"
	f.Method = "check_schedule.get"
	f.Params = map[string]uint64{"blah": 1}
	return f
}

func (r CheckScheduleGet) Encode() ([]byte, error) {
	return json.Marshal(r)
}

///////////////////////////////////////////////////////////////////////////////
// Poller Register

type PollerRegister struct {
	FrameMsg
	Params map[string][]string `json:"params"`
}

func NewPollerRegister(zones []string) Frame {
	f := &PollerRegister{}
	f.Version = "1"
	f.Method = "poller.register"
	f.Params = map[string][]string{"zones": zones}
	return f
}

func (r PollerRegister) Encode() ([]byte, error) {
	return json.Marshal(r)
}

///////////////////////////////////////////////////////////////////////////////
// HostInfo

type HostInfoResponse struct {
	FrameMsgCommon

	Result interface{} `json:"result"`
}

func NewHostInfoResponse(cr *check.CheckResult, f *FrameMsg, hinfo hostinfo.HostInfo) *HostInfoResponse {
	resp := &HostInfoResponse{}
	resp.Result = hinfo.BuildResult(cr)
	resp.SetResponseFrameMsg(f)

	return resp
}

func (r *HostInfoResponse) Encode() ([]byte, error) {
	return json.Marshal(r)
}

///////////////////////////////////////////////////////////////////////////////
// Metrics Post

type MetricWrap []map[string]*MetricTVU
type MetricWrapper []MetricWrap

func ConvertToMetricResults(crs *check.CheckResultSet) MetricWrap {
	wrappers := make(MetricWrap, 0)
	wrappers = append(wrappers, nil) // needed for the current protocol
	for i := 0; i < crs.Length(); i++ {
		cr := crs.Get(i)
		mapper := make(map[string]*MetricTVU)
		for key, m := range cr.Metrics {
			mapper[key] = &MetricTVU{
				Type:  m.TypeString,
				Value: fmt.Sprintf("%v", m.Value),
				Unit:  m.Unit,
			}
		}
		wrappers = append(wrappers, mapper)
	}
	return wrappers
}

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

func NewMetricsPostRequest(crs *check.CheckResultSet) *MetricsPostRequest {
	req := &MetricsPostRequest{}
	req.Version = "1"
	req.Method = "check_metrics.post"
	req.Params.EntityId = crs.Check.GetEntityId()
	req.Params.CheckId = crs.Check.GetId()
	req.Params.CheckType = crs.Check.GetCheckType()
	req.Params.Metrics = []MetricWrap{ConvertToMetricResults(crs)}
	req.Params.MinCheckPeriod = crs.Check.GetPeriod() * 1000
	req.Params.State = crs.State
	req.Params.Status = crs.Status
	req.Params.Timestamp = utils.NowTimestampMillis()
	return req
}

func (r MetricsPostRequest) Encode() ([]byte, error) {
	return json.Marshal(r)
}
