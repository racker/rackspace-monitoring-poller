//
// Copyright 2017 Rackspace
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

package check_test

import (
	"context"
	"errors"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewMetricsPostRequest(t *testing.T) {
	checkData := `{
	  "check_id":"chTestTCP_TLSRunSuccess",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"port":443,"ssl":true},
	  "type":"remote.tcp",
	  "timeout":15,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"IPv4",
	  "disabled":false
	  }`
	ch, err := check.NewCheck(context.Background(), []byte(checkData))
	assert.Equal(t, "chTestTCP_TLSRunSuccess", ch.GetID())
	require.NoError(t, err)

	cr1 := check.NewResult(metric.NewMetric("tt_connect", "", metric.MetricNumber, 250, metric.UnitMilliseconds))
	crs := check.NewResultSet(ch, cr1)
	// contrived, but add a distinct cr to validate proper JSON encoding
	cr2 := check.NewResult(metric.NewMetric("duration", "", metric.MetricNumber, 500, metric.UnitMilliseconds))
	crs.Add(cr2)

	origTimeStamper := utils.InstallAlternateTimestampFunc(func() int64 {
		return 5000
	})
	defer utils.InstallAlternateTimestampFunc(origTimeStamper)

	req := check.NewMetricsPostRequest(crs, 200)

	assert.Equal(t, int64(5200), req.Params.Timestamp)

	assert.Len(t, req.Params.Metrics, 2)

	assert.Len(t, req.Params.Metrics[0], 2)
	assert.Nil(t, req.Params.Metrics[0][0])
	assert.Equal(t, "250", req.Params.Metrics[0][1]["tt_connect"].Value)

	assert.Len(t, req.Params.Metrics[1], 2)
	assert.Nil(t, req.Params.Metrics[1][0])
	assert.Equal(t, "500", req.Params.Metrics[1][1]["duration"].Value)

	raw, err := req.Encode()
	require.NoError(t, err)
	assert.Equal(t, "{\"v\":\"1\",\"id\":0,\"target\":\"\",\"source\":\"\",\"method\":\"check_metrics.post_multi\","+
		"\"params\":{"+
		"\"entity_id\":\"enAAAAIPV4\",\"check_id\":\"chTestTCP_TLSRunSuccess\",\"check_type\":\"remote.tcp\","+
		"\"metrics\":["+
		"[null,{\"tt_connect\":{\"t\":\"int64\",\"v\":\"250\",\"u\":\"MILLISECONDS\"}}],"+
		"[null,{\"duration\":{\"t\":\"int64\",\"v\":\"500\",\"u\":\"MILLISECONDS\"}}]"+
		"],"+
		"\"min_check_period\":30000,\"state\":\"unavailable\",\"status\":\"unknown error\",\"timestamp\":5200}"+
		"}", string(raw))
}

func TestStates_SetStatusFromError(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		expected string
	}{
		{name: "with", in: "write ip6 ::->2001:4800:7902:1:0:a:4323:44: sendto: no route to host", expected: "no route to host"},
		{name: "without", in: "operation not permitted", expected: "operation not permitted"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var states check.States

			states.SetStatusFromError(errors.New(tt.in))
			assert.Equal(t, tt.expected, states.Status)
		})
	}
}
