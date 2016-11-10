package hostinfo

import (
	"encoding/json"
	"github.com/racker/rackspace-monitoring-poller/check"
)

type HostInfo interface {
	Run() (*check.CheckResult, error)
	BuildResult(cr *check.CheckResult) interface{}
}

type HostInfoBase struct {
	Type string `json:"type"`
}

func NewHostInfo(rawParams json.RawMessage) HostInfo {
	hinfo := &HostInfoBase{}
	err := json.Unmarshal(rawParams, &hinfo)
	if err != nil {
		return nil
	}
	switch hinfo.Type {
	case "MEMORY":
		return NewHostInfoMemory(hinfo)
	case "CPU":
		return NewHostInfoCpu(hinfo)
	default:
		return nil
	}
}
