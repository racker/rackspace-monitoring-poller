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

func (*HostInfoFilesystem) Run() (interface{}, error) {
	log.Debug("Running Filesystem")
	result := &hostinfo.HostInfoFilesystemResult{}
	result.Timestamp = utils.NowTimestampMillis()
	partitions, _ := disk.Partitions(false)
	for _, part := range partitions {
		if usage, err := disk.Usage(part.Mountpoint); err == nil {
			metrics := hostinfo.HostInfoFilesystemMetrics{}
			metrics.DirectoryName = part.Mountpoint
			metrics.DeviceName = part.Device
			metrics.SystemTypeName = part.Fstype
			metrics.Options = part.Opts
			metrics.Total = usage.Total / BytesToKilobytes
			metrics.Free = usage.Free / BytesToKilobytes
			metrics.Used = usage.Used / BytesToKilobytes
			metrics.Available = usage.Free / BytesToKilobytes
			metrics.Files = usage.InodesUsed
			metrics.FreeFiles = usage.InodesFree
			result.Metrics = append(result.Metrics, metrics)
		}
	}
	return result, nil
}
