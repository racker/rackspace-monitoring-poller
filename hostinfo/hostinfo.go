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

// HostInfo
package hostinfo

import (
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/protocol"
)

type HostInfo interface {
	Run() (*check.CheckResult, error)
	BuildResult(cr *check.CheckResult) interface{}
}

func NewHostInfoResponse(cr *check.CheckResult, f *protocol.FrameMsg, hinfo HostInfo) *protocol.HostInfoResponse {
	resp := &protocol.HostInfoResponse{}
	resp.Result = hinfo.BuildResult(cr)
	resp.SetResponseFrameMsg(f)

	return resp
}

func NewHostInfoResponse(cr *check.CheckResult, f *protocol.FrameMsg, hinfo HostInfo) *protocol.HostInfoResponse {
	resp := &protocol.HostInfoResponse{}
	resp.Result = hinfo.BuildResult(cr)
	resp.SetResponseFrameMsg(f)

	return resp
}
