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
	"bytes"
	"fmt"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	hostinfo_proto "github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"testing"
)

func TestHostInfoMemory_PopulateResult(t *testing.T) {
	hinfo := &hostinfo_proto.HostInfoBase{Type: "MEMORY"}
	hostInfoMemory := hostinfo.NewHostInfoMemory(hinfo)

	crs := check.NewCheckResultSet(nil, nil)
	cr := check.NewCheckResult()
	cr.AddMetrics(
		metric.NewMetric("UsedPercentage", "", metric.MetricFloat, 0.75, ""),
		metric.NewMetric("Free", "", metric.MetricNumber, 250, ""),
		metric.NewMetric("Total", "", metric.MetricNumber, 1000, ""),
		metric.NewMetric("Used", "", metric.MetricNumber, 750, ""),
		metric.NewMetric("SwapFree", "", metric.MetricNumber, 50, ""),
		metric.NewMetric("SwapTotal", "", metric.MetricNumber, 200, ""),
		metric.NewMetric("SwapUsed", "", metric.MetricNumber, 150, ""),
		metric.NewMetric("SwapUsedPercentage", "", metric.MetricFloat, 0.75, ""),
	)
	crs.Add(cr)

	sourceFrame := &protocol.FrameMsg{}
	utils.NowTimestampMillis = func() int64 { return 100 }
	response := hostinfo.NewHostInfoResponse(crs, sourceFrame, hostInfoMemory)
	encoded, err := response.Encode()
	if err != nil {
		t.Error(err)
	}

	expected := []byte(`{"v":"","id":0,"target":"","source":"","result":{"metrics":{"used_percentage":0.75,"actual_free":0,"actual_used":0,"free":0,"total":0,"used":0,"ram":0,"swap_free":0,"swap_total":0,"swap_used":0,"swap_percentage":0.75},"timestamp":100}}`)
	if !bytes.Equal(encoded, expected) {
		t.Error("wrong encoding")
	}
}

func TestHostInfoProcesses_PopulateResult(t *testing.T) {
	hinfo := &hostinfo_proto.HostInfoBase{Type: "PROCS"}
	hostInfoProcs := hostinfo.NewHostInfoProcesses(hinfo)
	crs, err := hostInfoProcs.Run()
	if err != nil {
		t.Error(err)
	}
	result := hostInfoProcs.BuildResult(crs)
	fmt.Printf("%v", result)
}
