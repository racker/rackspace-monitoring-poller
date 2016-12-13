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
	"github.com/shirou/gopsutil/process"
)

type HostInfoProcesses struct {
	hostinfo.HostInfoBase
}

///////////////////////////////////////////////////////////////////////////////
// HostInfo Memory

func NewHostInfoProcesses(base *hostinfo.HostInfoBase) HostInfo {
	return &HostInfoProcesses{HostInfoBase: *base}
}

func (*HostInfoProcesses) Run() (*check.CheckResultSet, error) {
	log.Debug("Running Processes")
	crs := check.NewCheckResultSet(nil, nil)
	pids, err := process.Pids()
	if err != nil {
		return nil, err
	}
	for _, pid := range pids {
		cr := check.NewCheckResult()
		pr, err := process.NewProcess(pid)
		if err != nil {
			continue
		}
		cr.AddMetric(metric.NewMetric("pid", "", metric.MetricNumber, pr.Pid, ""))
		if name, err := pr.Name(); err == nil {
			cr.AddMetric(
				metric.NewMetric("state_name", "", metric.MetricString, name, ""),
			)
		} else {
			continue
		}
		if cwd, err := pr.Cwd(); err == nil {
			cr.AddMetric(
				metric.NewMetric("exe_cwd", "", metric.MetricString, cwd, ""),
			)
		} else {
			continue
		}
		if root, err := pr.Exe(); err == nil {
			cr.AddMetric(
				metric.NewMetric("exe_root", "", metric.MetricString, root, ""),
			)
		} else {
			continue
		}
		if createTime, err := pr.CreateTime(); err == nil {
			cr.AddMetric(
				metric.NewMetric("time_start_time", "", metric.MetricNumber, createTime, ""),
			)
		} else {
			continue
		}
		if times, err := pr.Times(); err == nil {
			cr.AddMetrics(
				metric.NewMetric("time_user", "", metric.MetricFloat, times.User, ""),
				metric.NewMetric("time_sys", "", metric.MetricFloat, times.System, ""),
				metric.NewMetric("time_total", "", metric.MetricFloat, times.Total, ""),
			)
		} else {
			continue
		}
		if memory, err := pr.MemoryInfo(); err == nil {
			cr.AddMetrics(
				metric.NewMetric("memory_resident", "", metric.MetricNumber, memory.RSS, ""),
			)
		} else {
			continue
		}
		crs.Add(cr)
	}
	return crs, nil
}

func (hi *HostInfoProcesses) BuildResult(crs *check.CheckResultSet) interface{} {
	result := &hostinfo.HostInfoProcessesResult{}
	result.Timestamp = utils.NowTimestampMillis()
	for i := 0; i < crs.Length(); i++ {
		cr := crs.Get(i)
		metrics := hostinfo.HostInfoProcessesMetrics{}
		metrics.Pid, _ = cr.GetMetric("pid").ToInt32()
		metrics.StateName, _ = cr.GetMetric("state_name").ToString()
		metrics.ExeCwd, _ = cr.GetMetric("exe_cwd").ToString()
		metrics.ExeRoot, _ = cr.GetMetric("exe_root").ToString()
		metrics.StartTime, _ = cr.GetMetric("time_start_time").ToInt64()
		metrics.TimeUser, _ = cr.GetMetric("time_user").ToFloat64()
		metrics.TimeSys, _ = cr.GetMetric("time_sys").ToFloat64()
		metrics.TimeTotal, _ = cr.GetMetric("time_total").ToFloat64()
		metrics.MemoryRes, _ = cr.GetMetric("memory_resident").ToUint64()
		result.Metrics = append(result.Metrics, metrics)
	}
	return result
}
