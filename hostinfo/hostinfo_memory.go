package hostinfo

import (
	log "github.com/Sirupsen/logrus"
	"github.com/shirou/gopsutil/mem"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"
)

type HostInfoMemory struct {
	HostInfoBase
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

type HostInfoMemoryResult struct {
	Metrics   HostInfoMemoryMetrics `json:"metrics"`
	Timestamp int64                 `json:"timestamp"`
}

func NewHostInfoMemory(base *HostInfoBase) HostInfo {
	return &HostInfoMemory{HostInfoBase: *base}
}

func (*HostInfoMemory) Run() (*check.CheckResult, error) {
	log.Println("Running Memory")
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	s, err := mem.SwapMemory()
	if err != nil {
		return nil, err
	}
	cr := check.NewCheckResult()
	cr.AddMetrics(
		metric.NewMetric("UsedPercentage", "", metric.MetricFloat, v.UsedPercent, ""),
		metric.NewMetric("Free", "", metric.MetricNumber, v.Free, ""),
		metric.NewMetric("Total", "", metric.MetricNumber, v.Total, ""),
		metric.NewMetric("Used", "", metric.MetricNumber, v.Used, ""),
		metric.NewMetric("SwapFree", "", metric.MetricNumber, s.Free, ""),
		metric.NewMetric("SwapTotal", "", metric.MetricNumber, s.Total, ""),
		metric.NewMetric("SwapUsed", "", metric.MetricNumber, s.Used, ""),
		metric.NewMetric("SwapUsedPercentage", "", metric.MetricFloat, s.UsedPercent, ""),
	)
	return cr, nil
}

func (*HostInfoMemory) BuildResult(cr *check.CheckResult) interface{}  {
	result := &HostInfoMemoryResult{}
	
	result.Timestamp = utils.NowTimestampMillis()
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