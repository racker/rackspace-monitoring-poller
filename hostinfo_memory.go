package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/shirou/gopsutil/mem"
)

type HostInfoMemory struct {
	HostInfoBase
}

func NewHostInfoMemory(base *HostInfoBase) HostInfo {
	return &HostInfoMemory{HostInfoBase: *base}
}

func (*HostInfoMemory) Run() (*CheckResult, error) {
	log.Println("Running Memory")
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	s, err := mem.SwapMemory()
	if err != nil {
		return nil, err
	}
	cr := NewCheckResult()
	cr.AddMetrics(
		NewMetric("UsedPercentage", "", MetricFloat, v.UsedPercent, ""),
		NewMetric("Free", "", MetricNumber, v.Free, ""),
		NewMetric("Total", "", MetricNumber, v.Total, ""),
		NewMetric("Used", "", MetricNumber, v.Used, ""),
		NewMetric("SwapFree", "", MetricNumber, s.Free, ""),
		NewMetric("SwapTotal", "", MetricNumber, s.Total, ""),
		NewMetric("SwapUsed", "", MetricNumber, s.Used, ""),
		NewMetric("SwapUsedPercentage", "", MetricFloat, s.UsedPercent, ""),
	)
	return cr, nil
}
