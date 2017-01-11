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
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/shirou/gopsutil/cpu"
)

type HostInfoCpu struct {
	hostinfo.HostInfoBase
}

func NewHostInfoCpu(base *hostinfo.HostInfoBase) HostInfo {
	return &HostInfoCpu{HostInfoBase: *base}
}

func (*HostInfoCpu) Run() (*check.ResultSet, error) {
	log.Println("Running CPU")
	stats, err := cpu.Times(true)
	if err != nil {
		return nil, err
	}
	info, err := cpu.Info()
	if err != nil {
		return nil, err
	}
	coreCount, _ := cpu.Counts(true)
	crs := check.NewResultSet(nil, nil)
	for i, cpu := range info {
		cr := check.NewResult()
		cr.AddMetrics(
			metric.NewMetric("name", "", metric.MetricString, fmt.Sprintf("cpu.%d", i), ""),
			metric.NewMetric("model", "", metric.MetricString, cpu.ModelName, ""),
			metric.NewMetric("vendor", "", metric.MetricString, cpu.VendorID, ""),
			metric.NewMetric("idle", "", metric.MetricFloat, stats[i].Idle, ""),
			metric.NewMetric("irq", "", metric.MetricFloat, stats[i].Irq, ""),
			metric.NewMetric("mhz", "", metric.MetricFloat, cpu.Mhz, ""),
			metric.NewMetric("nice", "", metric.MetricFloat, stats[i].Nice, ""),
			metric.NewMetric("soft_irq", "", metric.MetricFloat, stats[i].Softirq, ""),
			metric.NewMetric("stolen", "", metric.MetricFloat, stats[i].Stolen, ""),
			metric.NewMetric("sys", "", metric.MetricFloat, stats[i].System, ""),
			metric.NewMetric("total", "", metric.MetricFloat, stats[i].Total(), ""),
			metric.NewMetric("total_cores", "", metric.MetricNumber, coreCount, ""),
			metric.NewMetric("total_sockets", "", metric.MetricNumber, uint64(len(info)), ""),
			metric.NewMetric("user", "", metric.MetricFloat, stats[i].User, ""),
			metric.NewMetric("wait", "", metric.MetricFloat, stats[i].Iowait, ""),
		)
		crs.Add(cr)
	}
	return crs, nil
}

func (*HostInfoCpu) BuildResult(crs *check.ResultSet) interface{} {
	result := &hostinfo.HostInfoCpuResult{}
	result.Timestamp = utils.NowTimestampMillis()
	if crs == nil {
		log.Infoln("Check result set is unset")
		return result
	}
	for i := 0; i < crs.Length(); i++ {
		cr := crs.Get(i)
		metrics := hostinfo.HostInfoCpuMetrics{}
		metrics.Name, _ = cr.GetMetric("name").ToString()
		metrics.Vendor, _ = cr.GetMetric("vendor").ToString()
		metrics.Model, _ = cr.GetMetric("model").ToString()
		metrics.SoftIrq, _ = cr.GetMetric("soft_irq").ToFloat64()
		metrics.User, _ = cr.GetMetric("user").ToFloat64()
		metrics.TotalSockets, _ = cr.GetMetric("total_sockets").ToUint64()
		metrics.TotalCores, _ = cr.GetMetric("total_cores").ToUint64()
		metrics.Nice, _ = cr.GetMetric("nice").ToFloat64()
		metrics.Wait, _ = cr.GetMetric("wait").ToFloat64()
		idle, _ := cr.GetMetric("idle").ToFloat64()
		metrics.Idle = uint64(idle)
		total, _ := cr.GetMetric("total").ToFloat64()
		metrics.Total = uint64(total)
		metrics.Stolen, _ = cr.GetMetric("stolen").ToFloat64()
		metrics.Irq, _ = cr.GetMetric("irq").ToFloat64()
		metrics.Sys, _ = cr.GetMetric("sys").ToFloat64()
		metrics.Mhz, _ = cr.GetMetric("mhz").ToFloat64()
		result.Metrics = append(result.Metrics, metrics)
	}
	return result
}
