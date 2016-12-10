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
	"github.com/shirou/gopsutil/host"
	"runtime"
)

type HostInfoSystem struct {
	hostinfo.HostInfoBase
}

///////////////////////////////////////////////////////////////////////////////
// HostInfo Memory

func NewHostInfoSystem(base *hostinfo.HostInfoBase) HostInfo {
	return &HostInfoSystem{HostInfoBase: *base}
}

func (*HostInfoSystem) Run() (*check.CheckResultSet, error) {
	log.Debug("Running System")
	info, _ := host.Info()
	crs := check.NewCheckResultSet(nil, nil)
	cr := check.NewCheckResult()
	cr.AddMetrics(
		metric.NewMetric("arch", "", metric.MetricString, runtime.GOOS, ""),
		metric.NewMetric("name", "", metric.MetricString, info.OS, ""),
		metric.NewMetric("version", "", metric.MetricString, info.KernelVersion, ""),
		metric.NewMetric("vendor_name", "", metric.MetricString, info.OS, ""),
		metric.NewMetric("vendor", "", metric.MetricString, info.Platform, ""),
		metric.NewMetric("vendor_version", "", metric.MetricString, info.PlatformVersion, ""),
	)
	crs.Add(cr)
	return crs, nil
}

func (*HostInfoSystem) BuildResult(crs *check.CheckResultSet) interface{} {
	result := &hostinfo.HostInfoSystemResult{}
	cr := crs.Get(0)
	result.Metrics.Arch, _ = cr.GetMetric("arch").ToString()
	result.Metrics.Name, _ = cr.GetMetric("name").ToString()
	result.Metrics.Version, _ = cr.GetMetric("version").ToString()
	result.Metrics.VendorName, _ = cr.GetMetric("vendor_name").ToString()
	result.Metrics.Vendor, _ = cr.GetMetric("vendor").ToString()
	result.Metrics.VendorVersion, _ = cr.GetMetric("vendor_version").ToString()
	return result
}
