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
	"github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
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

func (*HostInfoProcesses) Run() (interface{}, error) {
	log.Debug("Running Processes")
	result := &hostinfo.HostInfoProcessesResult{}
	result.Timestamp = utils.NowTimestampMillis()
	pids, err := process.Pids()
	if err != nil {
		return nil, err
	}
	for _, pid := range pids {
		metrics := hostinfo.HostInfoProcessesMetrics{}
		pr, err := process.NewProcess(pid)
		if err != nil {
			continue
		}
		metrics.Pid = pr.Pid
		if name, err := pr.Name(); err == nil {
			metrics.StateName = name
		} else {
			continue
		}
		if cwd, err := pr.Cwd(); err == nil {
			metrics.ExeCwd = cwd
		} else {
			continue
		}
		if root, err := pr.Exe(); err == nil {
			metrics.ExeRoot = root
		} else {
			continue
		}
		if createTime, err := pr.CreateTime(); err == nil {
			metrics.StartTime = createTime
		} else {
			continue
		}
		if times, err := pr.Times(); err == nil {
			metrics.TimeUser = times.User
			metrics.TimeSys = times.System
			metrics.TimeTotal = times.Total()
		} else {
			continue
		}
		if memory, err := pr.MemoryInfo(); err == nil {
			metrics.MemoryRes = memory.RSS
		} else {
			continue
		}
		result.Metrics = append(result.Metrics, metrics)
	}
	return result, nil
}
