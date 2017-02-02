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

// Hostinfo Test
package hostinfo_test

import (
	"github.com/racker/rackspace-monitoring-poller/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	hostinfo_proto "github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHostInfoMemory_PopulateResult(t *testing.T) {
	hinfo := &hostinfo_proto.HostInfoBase{Type: "MEMORY"}
	hostInfoMemory := hostinfo.NewHostInfoMemory(hinfo)
	result, err := hostInfoMemory.Run()
	sourceFrame := &protocol.FrameMsg{}
	utils.NowTimestampMillis = func() int64 { return 100 }
	response := hostinfo.NewHostInfoResponse(result, sourceFrame)
	encoded, err := response.Encode()
	assert.NoError(t, err)
	assert.NotEmpty(t, encoded)
}

func TestHostInfoProcesses_PopulateResult(t *testing.T) {
	hinfo := &hostinfo_proto.HostInfoBase{Type: "PROCS"}
	hostInfoProcs := hostinfo.NewHostInfoProcesses(hinfo)
	_, err := hostInfoProcs.Run()
	if err != nil {
		t.Error(err)
	}
}
