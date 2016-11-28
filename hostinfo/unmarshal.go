package hostinfo

import (
	"encoding/json"
	protocol "github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
)

func NewHostInfo(rawParams json.RawMessage) HostInfo {
	hinfo := &protocol.HostInfoBase{}

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
