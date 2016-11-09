package main

import (
	"encoding/json"
)

type HostInfo interface {
	Run() (*CheckResult, error)
}

type HostInfoBase struct {
	Type string `json:"type"`
}

func NewHostInfo(f Frame) HostInfo {
	hinfo := &HostInfoBase{}
	err := json.Unmarshal(*f.GetRawParams(), &hinfo)
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
