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

func (*HostInfoMemory) BuildResult(cr *CheckResult) interface{}  {
	result := &HostInfoMemoryResult{}
	
	result.Timestamp = NowTimestampMillis()
	result.Metrics.UsedPercentage, _ = cr.GetMetric("UsedPercentage").ToFloat64()
	result.Metrics.Free, _ = cr.GetMetric("Free").ToUint64()
	result.Metrics.Total, _ = cr.GetMetric("Total").ToUint64()
	result.Metrics.Used, _ = cr.GetMetric("Used").ToUint64()
	result.Metrics.SwapFree, _ = cr.GetMetric("UsedPercentage").ToUint64()
	result.Metrics.SwapTotal, _ = cr.GetMetric("SwapTotal").ToUint64()
	result.Metrics.SwapUsed, _ = cr.GetMetric("SwapUsed").ToUint64()
	result.Metrics.SwapUsedPercentage, _ = cr.GetMetric("SwapUsedPercentage").ToFloat64()

	return result
}