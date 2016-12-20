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
	"github.com/shirou/gopsutil/disk"
)

type HostInfoFilesystem struct {
	hostinfo.HostInfoBase
}

const BytesToKilobytes = 1024

///////////////////////////////////////////////////////////////////////////////
// HostInfo Memory

func NewHostInfoFilesystem(base *hostinfo.HostInfoBase) HostInfo {
	return &HostInfoFilesystem{HostInfoBase: *base}
}

func (*HostInfoFilesystem) Run() (*check.CheckResultSet, error) {
	log.Debug("Running Filesystem")
	crs := check.NewCheckResultSet(nil, nil)
	partitions, _ := disk.Partitions(false)
	for _, part := range partitions {
		if usage, err := disk.Usage(part.Mountpoint); err == nil {
			cr := check.NewCheckResult()
			cr.AddMetrics(
				metric.NewMetric("dir_name", "", metric.MetricString, part.Mountpoint, ""),
				metric.NewMetric("dev_name", "", metric.MetricString, part.Device, ""),
				metric.NewMetric("sys_type_name", "", metric.MetricString, part.Fstype, ""),
				metric.NewMetric("options", "", metric.MetricString, part.Opts, ""),

				metric.NewMetric("total", "", metric.MetricNumber, usage.Total/BytesToKilobytes, ""),
				metric.NewMetric("free", "", metric.MetricNumber, usage.Free/BytesToKilobytes, ""),
				metric.NewMetric("used", "", metric.MetricNumber, usage.Used/BytesToKilobytes, ""),
				metric.NewMetric("avail", "", metric.MetricNumber, usage.Free/BytesToKilobytes, ""),
				metric.NewMetric("files", "", metric.MetricNumber, usage.InodesUsed, ""),
				metric.NewMetric("free_files", "", metric.MetricNumber, usage.InodesFree, ""),
			)
			crs.Add(cr)
		}
	}
	return crs, nil
}

func (*HostInfoFilesystem) BuildResult(crs *check.CheckResultSet) interface{} {
	result := &hostinfo.HostInfoFilesystemResult{}
	result.Timestamp = utils.NowTimestampMillis()
	if crs == nil {
		log.Infoln("Check result set is unset")
		return result
	}
	for i := 0; i < crs.Length(); i++ {
		cr := crs.Get(i)
		metrics := hostinfo.HostInfoFilesystemMetrics{}
		metrics.DirectoryName, _ = cr.GetMetric("dir_name").ToString()
		metrics.DeviceName, _ = cr.GetMetric("dev_name").ToString()
		metrics.SystemTypeName, _ = cr.GetMetric("sys_type_name").ToString()
		metrics.Options, _ = cr.GetMetric("options").ToString()

		metrics.Total, _ = cr.GetMetric("total").ToUint64()
		metrics.Free, _ = cr.GetMetric("free").ToUint64()
		metrics.Used, _ = cr.GetMetric("used").ToUint64()
		metrics.Available, _ = cr.GetMetric("avail").ToUint64()
		metrics.Files, _ = cr.GetMetric("files").ToUint64()
		metrics.FreeFiles, _ = cr.GetMetric("free_files").ToUint64()
		result.Metrics = append(result.Metrics, metrics)
	}
	return result
}
