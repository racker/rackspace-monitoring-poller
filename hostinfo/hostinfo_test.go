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
package hostinfo_test

import (
	"bytes"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/metric"
	"github.com/racker/rackspace-monitoring-poller/types"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"log"
	"testing"
)

func TestHostInfoMemory_PopulateResult(t *testing.T) {
	hinfo := &hostinfo.HostInfoBase{Type: "MEMORY"}
	hostInfoMemory := hostinfo.NewHostInfoMemory(hinfo)

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

	sourceFrame := &types.FrameMsg{}

	utils.NowTimestampMillis = func() int64 { return 100 }

	response := &types.HostInfoResponse{}
	response.Result = hostInfoMemory.BuildResult(cr)
	response.SetResponseFrameMsg(sourceFrame)

	encoded, err := response.Encode()
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(encoded, []byte("{\"v\":\"\",\"id\":0,\"target\":\"\",\"source\":\"\",\"result\":{\"metrics\":{\"used_percentage\":0.75,\"free\":0,\"total\":0,\"used\":0,\"swap_free\":0,\"swap_total\":0,\"swap_used\":0,\"swap_percentage\":0.75},\"timestamp\":100}}")) {
		log.Printf(string(encoded))
		t.Error("wrong encoding")
	}
}
