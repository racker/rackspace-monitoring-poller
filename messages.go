package main

import (
	"encoding/json"
	"fmt"
	"time"
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

func NowTimestampMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
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
	f.Params.Timestamp = NowTimestampMillis()
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
// HostInfo Memory

type HostInfoMemoryMetrics struct {
	UsedPercentage     float64 `json:"used_percentage"`
	Free               uint64  `json:"free"`
	Total              uint64  `json:"total"`
	Used               uint64  `json:"used"`
	SwapFree           uint64  `json:"swap_free"`
	SwapTotal          uint64  `json:"swap_total"`
	SwapUsed           uint64  `json:"swap_used"`
	SwapUsedPercentage float64 `json:"swap_percentage"`
}

type HostInfoMemoryResponse struct {
	FrameMsg
	Result struct {
		Metrics   HostInfoMemoryMetrics `json:"metrics"`
		Timestamp int64                 `json:"timestamp"`
	} `json:"result"`
}

func NewHostInfoResponse(frame *FrameMsg, hinfo HostInfo, cr *CheckResult) Frame {
	resp := &HostInfoMemoryResponse{}
	resp.SetResponseFrameMsg(frame)
	resp.Result.Timestamp = NowTimestampMillis()
	resp.Result.Metrics.UsedPercentage, _ = cr.GetMetric("UsedPercentage").ToFloat64()
	resp.Result.Metrics.Free, _ = cr.GetMetric("Free").ToUint64()
	resp.Result.Metrics.Total, _ = cr.GetMetric("Total").ToUint64()
	resp.Result.Metrics.Used, _ = cr.GetMetric("Used").ToUint64()
	resp.Result.Metrics.SwapFree, _ = cr.GetMetric("UsedPercentage").ToUint64()
	resp.Result.Metrics.SwapTotal, _ = cr.GetMetric("SwapTotal").ToUint64()
	resp.Result.Metrics.SwapUsed, _ = cr.GetMetric("SwapUsed").ToUint64()
	resp.Result.Metrics.SwapUsedPercentage, _ = cr.GetMetric("SwapUsedPercentage").ToFloat64()
	return resp
}

func (r HostInfoMemoryResponse) Encode() ([]byte, error) {
	return json.Marshal(r)
}

///////////////////////////////////////////////////////////////////////////////
// Metrics Post

type MetricWrap []map[string]*MetricTVU
type MetricWrapper []MetricWrap

func ConvertToMetricResults(crs *CheckResultSet) MetricWrap {
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

func NewMetricsPostRequest(crs *CheckResultSet) *MetricsPostRequest {
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
	req.Params.Timestamp = NowTimestampMillis()
	return req
}

func (r MetricsPostRequest) Encode() ([]byte, error) {
	return json.Marshal(r)
}
