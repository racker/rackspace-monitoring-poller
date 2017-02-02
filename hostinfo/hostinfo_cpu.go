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
	"github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/shirou/gopsutil/cpu"
)

type HostInfoCpu struct {
	hostinfo.HostInfoBase
}

func NewHostInfoCpu(base *hostinfo.HostInfoBase) HostInfo {
	return &HostInfoCpu{HostInfoBase: *base}
}

func (*HostInfoCpu) Run() (interface{}, error) {
	log.Println("Running CPU")
	result := &hostinfo.HostInfoCpuResult{}
	stats, err := cpu.Times(true)
	if err != nil {
		return nil, err
	}
	info, err := cpu.Info()
	if err != nil {
		return nil, err
	}
	coreCount, _ := cpu.Counts(true)
	for i, cpu := range info {
		metrics := hostinfo.HostInfoCpuMetrics{}
		metrics.Name = fmt.Sprintf("cpu.%d", i)
		metrics.Model = cpu.ModelName
		metrics.Vendor = cpu.VendorID
		metrics.Idle = uint64(stats[i].Idle)
		metrics.Irq = stats[i].Irq
		metrics.Mhz = cpu.Mhz
		metrics.Nice = stats[i].Nice
		metrics.SoftIrq = stats[i].Softirq
		metrics.Stolen = stats[i].Stolen
		metrics.Sys = stats[i].System
		metrics.Total = uint64(stats[i].Total())
		metrics.TotalCores = uint64(coreCount)
		metrics.TotalSockets = uint64(len(info))
		metrics.User = stats[i].User
		metrics.Wait = stats[i].Iowait
		result.Metrics = append(result.Metrics, metrics)
	}
	result.Timestamp = utils.NowTimestampMillis()
	return result, nil
}
