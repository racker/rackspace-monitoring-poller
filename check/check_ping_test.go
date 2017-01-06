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
	"github.com/sparrc/go-ping"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const checkDataTemplate = `{
	  "id":"chPzATCP",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"count":%d},
	  "type":"remote.ping",
	  "timeout":%d,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"",
	  "disabled":false
	  }`

func setup(ctrl *gomock.Controller) *check.MockPinger {
	mockPinger := check.NewMockPinger(ctrl)
	check.PingerFactory = func(addr string) (check.Pinger, error) {
		return mockPinger, nil
	}
	return mockPinger
}

func TestPingCheck_ConfirmType(t *testing.T) {
	const count = 5
	const timeout = 15

	checkData := fmt.Sprintf(checkDataTemplate, count, timeout)
	c := check.NewCheck([]byte(checkData), context.Background(), func() {})

	assert.IsType(t, &check.PingCheck{}, c)
}

func TestPingCheck_PingerConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := setup(ctrl)

	const count = 5
	const timeout = 15

	mock.EXPECT().SetCount(count)
	mock.EXPECT().Count().Return(count)
	mock.EXPECT().SetTimeout(timeout * time.Second)
	mock.EXPECT().Timeout().Return(timeout * time.Second)
	mock.EXPECT().SetOnRecv(gomock.Any())
	mock.EXPECT().Run()
	statistics := &ping.Statistics{}
	mock.EXPECT().Statistics().Return(statistics)

	checkData := fmt.Sprintf(checkDataTemplate, count, timeout)
	c := check.NewCheck([]byte(checkData), context.Background(), func() {})

	// Run check since need to induce pinger usage
	crs, err := c.Run()
	assert.NotNil(t, crs)
	assert.NoError(t, err)
}

func TestPingCheck_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := setup(ctrl)

	const count = 5
	const timeout = 15

	mock.EXPECT().SetCount(count)
	mock.EXPECT().Count().Return(count)
	mock.EXPECT().SetTimeout(timeout * time.Second)
	mock.EXPECT().Timeout().Return(timeout * time.Second)
	mock.EXPECT().SetOnRecv(gomock.Any())
	mock.EXPECT().Run()
	statistics := &ping.Statistics{
		PacketsSent: count,
		PacketsRecv: count,
		MinRtt:      1 * time.Millisecond,
		AvgRtt:      5 * time.Millisecond,
		MaxRtt:      10 * time.Millisecond,
	}
	mock.EXPECT().Statistics().Return(statistics)

	checkData := fmt.Sprintf(checkDataTemplate, count, timeout)
	c := check.NewCheck([]byte(checkData), context.Background(), func() {})

	// Run check
	crs, err := c.Run()
	require.NoError(t, err)

	expected := []*ExpectedMetric{
		ExpectMetric("average", "", metric.MetricFloat, 0.005, metric.UnitSeconds),
		ExpectMetric("maximum", "", metric.MetricFloat, 0.010, metric.UnitSeconds),
		ExpectMetric("minimum", "", metric.MetricFloat, 0.001, metric.UnitSeconds),
		ExpectMetric("available", "", metric.MetricFloat, 100.0, metric.UnitPercent),
		ExpectMetric("count", "", metric.MetricNumber, count, ""),
	}

	assert.Equal(t, 1, crs.Length())
	AssertMetrics(t, expected, crs.Get(0).Metrics)
	assert.True(t, crs.Available)
}

func TestPingCheck_Drops(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := setup(ctrl)

	const count = 5
	const timeout = 15

	mock.EXPECT().SetCount(count)
	mock.EXPECT().Count().Return(count)
	mock.EXPECT().SetTimeout(timeout * time.Second)
	mock.EXPECT().Timeout().Return(timeout * time.Second)
	mock.EXPECT().SetOnRecv(gomock.Any())
	mock.EXPECT().Run()
	statistics := &ping.Statistics{
		PacketsSent: 2, // pretend a few trickled through, but timed out
		PacketsRecv: 0, // ...and none came back
		// leave others at defaults
	}
	mock.EXPECT().Statistics().Return(statistics)

	checkData := fmt.Sprintf(checkDataTemplate, count, timeout)
	c := check.NewCheck([]byte(checkData), context.Background(), func() {})

	assert.IsType(t, &check.PingCheck{}, c)

	// Run check
	crs, err := c.Run()
	require.NoError(t, err)

	expected := []*ExpectedMetric{
		ExpectMetric("average", "", metric.MetricFloat, 0.0, metric.UnitSeconds),
		ExpectMetric("maximum", "", metric.MetricFloat, 0.0, metric.UnitSeconds),
		ExpectMetric("minimum", "", metric.MetricFloat, 0.0, metric.UnitSeconds),
		ExpectMetric("available", "", metric.MetricFloat, 0.0, metric.UnitPercent),
		ExpectMetric("count", "", metric.MetricNumber, 2, ""),
	}

	assert.Equal(t, 1, crs.Length())
	AssertMetrics(t, expected, crs.Get(0).Metrics)
	assert.True(t, crs.Available)
}

func TestPingCheck_NoPermissionToPing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := setup(ctrl)

	const count = 5
	const timeout = 15

	mock.EXPECT().SetCount(count)
	mock.EXPECT().Count().Return(count)
	mock.EXPECT().SetTimeout(timeout * time.Second)
	mock.EXPECT().Timeout().Return(timeout * time.Second)
	mock.EXPECT().SetOnRecv(gomock.Any())
	mock.EXPECT().Run()
	statistics := &ping.Statistics{
		PacketsSent: 0, // go-ping doesn't emit any error, just logs it
		// leave others at defaults
	}
	mock.EXPECT().Statistics().Return(statistics)

	checkData := fmt.Sprintf(checkDataTemplate, count, timeout)
	c := check.NewCheck([]byte(checkData), context.Background(), func() {})

	assert.IsType(t, &check.PingCheck{}, c)

	// Run check
	crs, err := c.Run()
	require.NoError(t, err)

	assert.Equal(t, 0, crs.Length())
	assert.False(t, crs.Available)
}
