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
	log "github.com/sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/shirou/gopsutil/mem"
)

type HostInfoMemory struct {
	hostinfo.HostInfoBase
}

const Megabyte = 1024 * 1024

///////////////////////////////////////////////////////////////////////////////
// HostInfo Memory

func NewHostInfoMemory(base *hostinfo.HostInfoBase) HostInfo {
	return &HostInfoMemory{HostInfoBase: *base}
}

func (*HostInfoMemory) Run() (interface{}, error) {
	result := &hostinfo.HostInfoMemoryResult{}
	result.Timestamp = utils.NowTimestampMillis()
	log.Debug("Running Memory")
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	s, err := mem.SwapMemory()
	if err != nil {
		return nil, err
	}
	result.Metrics.UsedPercentage = v.UsedPercent
	result.Metrics.Free = v.Free
	result.Metrics.ActualFree = v.Free
	result.Metrics.Total = v.Total
	result.Metrics.Used = v.Used
	result.Metrics.ActualUsed = v.Used
	result.Metrics.SwapFree = s.Free
	result.Metrics.SwapTotal = s.Total
	result.Metrics.SwapUsed = s.Used
	result.Metrics.SwapUsedPercentage = s.UsedPercent
	result.Metrics.RAM = uint64(v.Total / Megabyte)
	return result, nil
}
