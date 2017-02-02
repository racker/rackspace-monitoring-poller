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
	"github.com/shirou/gopsutil/host"
	"runtime"
	"strings"
)

type HostInfoSystem struct {
	hostinfo.HostInfoBase
}

///////////////////////////////////////////////////////////////////////////////
// HostInfo Memory

func NewHostInfoSystem(base *hostinfo.HostInfoBase) HostInfo {
	return &HostInfoSystem{HostInfoBase: *base}
}

func (*HostInfoSystem) Run() (interface{}, error) {
	log.Debug("Running System")
	result := &hostinfo.HostInfoSystemResult{}
	info, _ := host.Info()
	os := strings.ToLower(info.OS)
	result.Metrics.Arch = runtime.GOOS
	result.Metrics.Name = os
	result.Metrics.Version = info.KernelVersion
	result.Metrics.VendorName = os
	result.Metrics.Vendor = info.Platform
	result.Metrics.VendorVersion = info.PlatformVersion
	return result, nil
}
