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
	"fmt"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPingCheck_success(t *testing.T) {
	const count = 5
	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chPzATCP",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"count":%d},
	  "type":"remote.ping",
	  "timeout":15,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"",
	  "disabled":false
	  }`, count)
	c := check.NewCheck([]byte(checkData))

	assert.IsType(t, &check.PingCheck{}, c)

	var crs *check.CheckResultSet
	var err error
	// Run check
	Timebox(t, 15*time.Second, func(t *testing.T) {
		crs, err = c.Run()
		require.NoError(t, err)
	})

	expected := []*ExpectedMetric{
		ExpectMetric("average", "", metric.MetricFloat, 0, metric.UnitSeconds).ButIgnoreValue(),
		ExpectMetric("maximum", "", metric.MetricFloat, 0, metric.UnitSeconds).ButIgnoreValue(),
		ExpectMetric("minimum", "", metric.MetricFloat, 0, metric.UnitSeconds).ButIgnoreValue(),
		ExpectMetric("available", "", metric.MetricFloat, 100.0, metric.UnitPercent),
		ExpectMetric("count", "", metric.MetricNumber, count, ""),
	}

	AssertMetrics(t, expected, crs.Get(0).Metrics)
}

func TestPingCheck_drops(t *testing.T) {
	const count = 5
	// Create Check
	// ...see RFC 5737 for background choosing 192.0.2.1 as a "non existent" IP address
	checkData := fmt.Sprintf(`{
	  "id":"chPzATCP",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"count":%d},
	  "type":"remote.ping",
	  "timeout":2,
	  "period":30,
	  "ip_addresses":{"default":"192.0.2.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"",
	  "disabled":false
	  }`, count)
	c := check.NewCheck([]byte(checkData))

	assert.IsType(t, &check.PingCheck{}, c)

	var crs *check.CheckResultSet
	var err error
	// Run check
	Timebox(t, 6*time.Second, func(t *testing.T) {
		crs, err = c.Run()
		require.NoError(t, err)
	})

	expected := []*ExpectedMetric{
		ExpectMetric("average", "", metric.MetricFloat, 0.0, metric.UnitSeconds),
		ExpectMetric("maximum", "", metric.MetricFloat, 0.0, metric.UnitSeconds),
		ExpectMetric("minimum", "", metric.MetricFloat, 0.0, metric.UnitSeconds),
		ExpectMetric("available", "", metric.MetricFloat, 0.0, metric.UnitPercent),
		ExpectMetric("count", "", metric.MetricNumber, 0, "").ButNonZeroValue(),
	}

	AssertMetrics(t, expected, crs.Get(0).Metrics)
}
