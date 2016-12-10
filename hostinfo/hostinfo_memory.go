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

package hostinfo

import (
	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/shirou/gopsutil/mem"
)

type HostInfoMemory struct {
	hostinfo.HostInfoBase
}

///////////////////////////////////////////////////////////////////////////////
// HostInfo Memory

func NewHostInfoMemory(base *hostinfo.HostInfoBase) HostInfo {
	return &HostInfoMemory{HostInfoBase: *base}
}

func (*HostInfoMemory) Run() (*check.CheckResultSet, error) {
	log.Debug("Running Memory")
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	s, err := mem.SwapMemory()
	if err != nil {
		return nil, err
	}
	crs := check.NewCheckResultSet(nil, nil)
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
	crs.Add(cr)
	return crs, nil
}

func (hi *HostInfoMemory) BuildResult(crs *check.CheckResultSet) interface{} {
	result := &hostinfo.HostInfoMemoryResult{}
	cr := crs.Get(0)
	result.Timestamp = utils.NowTimestampMillis()
	result.Metrics.UsedPercentage, _ = cr.GetMetric("UsedPercentage").ToFloat64()
	result.Metrics.Free, _ = cr.GetMetric("Free").ToUint64()
	result.Metrics.ActualFree, _ = cr.GetMetric("Free").ToUint64()
	result.Metrics.Total, _ = cr.GetMetric("Total").ToUint64()
	result.Metrics.Used, _ = cr.GetMetric("Used").ToUint64()
	result.Metrics.ActualUsed, _ = cr.GetMetric("Used").ToUint64()
	result.Metrics.SwapFree, _ = cr.GetMetric("SwapFree").ToUint64()
	result.Metrics.SwapTotal, _ = cr.GetMetric("SwapTotal").ToUint64()
	result.Metrics.SwapUsed, _ = cr.GetMetric("SwapUsed").ToUint64()
	result.Metrics.SwapUsedPercentage, _ = cr.GetMetric("SwapUsedPercentage").ToFloat64()
	ram, _ := cr.GetMetric("Total").ToUint64()
	result.Metrics.RAM = uint64(ram / (1024 * 1024))

	return result
}
