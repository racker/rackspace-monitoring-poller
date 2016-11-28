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

// Hostinfo CPU
package hostinfo

import (
	//"github.com/shirou/gopsutil/cpu"
	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
)

type HostInfoCpu struct {
	hostinfo.HostInfoBase
}

func NewHostInfoCpu(base *hostinfo.HostInfoBase) HostInfo {
	return &HostInfoCpu{HostInfoBase: *base}
}

func (*HostInfoCpu) Run() (*check.CheckResult, error) {
	log.Println("Running CPU")
	/*
		stats, err := cpu.Times(false)
		if err != nil {
			return nil, err
		}
		info, err := cpu.Info()
		if err != nil {
			return nil, err
		}
	*/
	cr := check.NewCheckResult()
	//TODO
	return cr, nil
}

func (*HostInfoCpu) BuildResult(cr *check.CheckResult) interface{} {
	//TODO
	return nil
}
