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
	"encoding/json"
	protocol "github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
)

func NewHostInfo(rawParams json.RawMessage) HostInfo {
	hinfo := &protocol.HostInfoBase{}

	err := json.Unmarshal(rawParams, &hinfo)
	if err != nil {
		return nil
	}
	switch hinfo.Type {
	case protocol.Memory:
		return NewHostInfoMemory(hinfo)
	case protocol.Cpu:
		return NewHostInfoCpu(hinfo)
	case protocol.Filesystem:
		return NewHostInfoFilesystem(hinfo)
	case protocol.System:
		return NewHostInfoSystem(hinfo)
	case protocol.Processes:
		return NewHostInfoProcesses(hinfo)
	default:
		return nil
	}
}
