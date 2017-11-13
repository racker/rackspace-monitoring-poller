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
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"sync"
)

const checkDataTemplate = `{
	  "id":"%s",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"count":%d},
	  "type":"remote.ping",
	  "timeout":%d,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"IPv4",
	  "disabled":false
	  }`

func TestPingCheck_ConfirmType(t *testing.T) {
	const count = 5
	const timeout = 15

	checkData := fmt.Sprintf(checkDataTemplate, "TestPingCheck_ConfirmType", count, timeout)
	c, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	assert.IsType(t, &check.PingCheck{}, c)
}

func TestPingCheck_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	const count = 5
	const timeout = 15

	checkData := fmt.Sprintf(checkDataTemplate, "TestPingCheck_Success", count, timeout)
	c, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	// Run check
	crs, err := c.Run()
	require.NoError(t, err)

	expected := []*ExpectedMetric{
		ExpectMetric("average", "", metric.MetricFloat, 0, metric.UnitSeconds).ButNonZeroValue(),
		ExpectMetric("maximum", "", metric.MetricFloat, 0, metric.UnitSeconds).ButNonZeroValue(),
		ExpectMetric("minimum", "", metric.MetricFloat, 0, metric.UnitSeconds).ButNonZeroValue(),
		ExpectMetric("available", "", metric.MetricFloat, 100.0, metric.UnitPercent),
		ExpectMetric("count", "", metric.MetricNumber, count, ""),
	}

	assert.Equal(t, 1, crs.Length())
	AssertMetrics(t, expected, crs.Get(0).Metrics)
	assert.True(t, crs.Available)
}

func TestPingCheck_Concurrent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	const count = 5
	const timeout = 15
	const concurrency = 50

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)

		go func(index int) {
			defer wg.Done()

			name := fmt.Sprintf("TestPingCheck_Concurrent-%d", index)

			t.Run(name, func(t *testing.T) {
				checkData := fmt.Sprintf(checkDataTemplate, name, count, timeout)

				c, err := check.NewCheck(context.Background(), []byte(checkData))
				require.NoError(t, err)

				// Run check
				crs, err := c.Run()
				require.NoError(t, err)

				expected := []*ExpectedMetric{
					ExpectMetric("average", "", metric.MetricFloat, 0, metric.UnitSeconds).ButNonZeroValue(),
					ExpectMetric("maximum", "", metric.MetricFloat, 0, metric.UnitSeconds).ButNonZeroValue(),
					ExpectMetric("minimum", "", metric.MetricFloat, 0, metric.UnitSeconds).ButNonZeroValue(),
					ExpectMetric("available", "", metric.MetricFloat, 100.0, metric.UnitPercent),
					ExpectMetric("count", "", metric.MetricNumber, count, ""),
				}

				assert.Equal(t, 1, crs.Length())
				AssertMetrics(t, expected, crs.Get(0).Metrics)
				assert.True(t, crs.Available)

			})

		}(i)
	}

	wg.Wait()

}
