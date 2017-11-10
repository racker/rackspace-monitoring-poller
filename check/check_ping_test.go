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
	"time"
	"github.com/pkg/errors"
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
	  "target_resolver":"IPv4",
	  "disabled":false
	  }`

func setup(ctrl *gomock.Controller) (*MockPinger, check.PingerFactorySpec) {
	mockPinger := NewMockPinger(ctrl)
	originalFactory := check.PingerFactory
	check.PingerFactory = func(identifier string, remoteAddr string, ipVersion string) (check.Pinger, error) {
		return mockPinger, nil
	}
	return mockPinger, originalFactory
}

func teardown(originalFactory check.PingerFactorySpec) {
	check.PingerFactory = originalFactory
}

func TestPingCheck_ConfirmType(t *testing.T) {
	const count = 5
	const timeout = 15

	checkData := fmt.Sprintf(checkDataTemplate, count, timeout)
	c, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	assert.IsType(t, &check.PingCheck{}, c)
}

func TestPingCheck_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock, originalFactory := setup(ctrl)
	defer teardown(originalFactory)

	const count = 5
	const timeout = 15

	mock.EXPECT().Ping(1, 3*time.Second).AnyTimes().Return(check.PingResponse{Seq: 1, Rtt: 1 * time.Millisecond})
	mock.EXPECT().Ping(2, 3*time.Second).AnyTimes().Return(check.PingResponse{Seq: 2, Rtt: 10 * time.Millisecond})
	mock.EXPECT().Ping(3, 3*time.Second).AnyTimes().Return(check.PingResponse{Seq: 3, Rtt: 5 * time.Millisecond})
	mock.EXPECT().Ping(4, 3*time.Second).AnyTimes().Return(check.PingResponse{Seq: 4, Rtt: 5 * time.Millisecond})
	mock.EXPECT().Ping(5, 3*time.Second).AnyTimes().Return(check.PingResponse{Seq: 5, Rtt: 5 * time.Millisecond})
	mock.EXPECT().Close()

	checkData := fmt.Sprintf(checkDataTemplate, count, timeout)
	c, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	// Run check
	crs, err := c.Run()
	require.NoError(t, err)

	expected := []*ExpectedMetric{
		ExpectMetric("average", "", metric.MetricFloat, 0.0052, metric.UnitSeconds),
		ExpectMetric("maximum", "", metric.MetricFloat, 0.010, metric.UnitSeconds),
		ExpectMetric("minimum", "", metric.MetricFloat, 0.001, metric.UnitSeconds),
		ExpectMetric("available", "", metric.MetricFloat, 100.0, metric.UnitPercent),
		ExpectMetric("count", "", metric.MetricNumber, count, ""),
	}

	assert.Equal(t, 1, crs.Length())
	AssertMetrics(t, expected, crs.Get(0).Metrics)
	assert.True(t, crs.Available)
}

func TestPingCheck_OutOfOrderResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock, originalFactory := setup(ctrl)
	defer teardown(originalFactory)

	const count = 5
	const timeout = 15

	// NOTE: the seq in the mocked response is "out of order"
	mock.EXPECT().Ping(1, 3*time.Second).AnyTimes().Return(check.PingResponse{Seq: 1, Rtt: 1 * time.Millisecond})
	mock.EXPECT().Ping(2, 3*time.Second).AnyTimes().Return(check.PingResponse{Seq: 3, Rtt: 5 * time.Millisecond})
	mock.EXPECT().Ping(3, 3*time.Second).AnyTimes().Return(check.PingResponse{Seq: 2, Rtt: 10 * time.Millisecond})
	mock.EXPECT().Ping(4, 3*time.Second).AnyTimes().Return(check.PingResponse{Seq: 5, Rtt: 5 * time.Millisecond})
	mock.EXPECT().Ping(5, 3*time.Second).AnyTimes().Return(check.PingResponse{Seq: 4, Rtt: 5 * time.Millisecond})
	mock.EXPECT().Close()

	checkData := fmt.Sprintf(checkDataTemplate, count, timeout)
	c, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	// Run check
	crs, err := c.Run()
	require.NoError(t, err)

	expected := []*ExpectedMetric{
		ExpectMetric("average", "", metric.MetricFloat, 0.0052, metric.UnitSeconds),
		ExpectMetric("maximum", "", metric.MetricFloat, 0.010, metric.UnitSeconds),
		ExpectMetric("minimum", "", metric.MetricFloat, 0.001, metric.UnitSeconds),
		ExpectMetric("available", "", metric.MetricFloat, 100.0, metric.UnitPercent),
		ExpectMetric("count", "", metric.MetricNumber, count, ""),
	}

	assert.Equal(t, 1, crs.Length())
	AssertMetrics(t, expected, crs.Get(0).Metrics)
	assert.True(t, crs.Available)
}

func TestPingCheck_PartialResponses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock, originalFactory := setup(ctrl)
	defer teardown(originalFactory)

	const count = 5
	const timeout = 15

	// NOTE: the seq in the mocked response is "out of order"
	mock.EXPECT().Ping(1, 3*time.Second).AnyTimes().Return(check.PingResponse{Seq: 1, Rtt: 2 * time.Millisecond})
	mock.EXPECT().Ping(2, 3*time.Second).AnyTimes().Return(check.PingResponse{Seq: 3, Rtt: 5 * time.Millisecond})
	mock.EXPECT().Ping(3, 3*time.Second).AnyTimes().Return(check.PingResponse{Err: errors.New("timeout"), Timeout: true})
	mock.EXPECT().Ping(4, 3*time.Second).AnyTimes().Return(check.PingResponse{Seq: 5, Rtt: 5 * time.Millisecond})
	mock.EXPECT().Ping(5, 3*time.Second).AnyTimes().Return(check.PingResponse{Err: errors.New("timeout"), Timeout: true})
	mock.EXPECT().Close()

	checkData := fmt.Sprintf(checkDataTemplate, count, timeout)
	c, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	// Run check
	crs, err := c.Run()
	require.NoError(t, err)

	expected := []*ExpectedMetric{
		ExpectMetric("average", "", metric.MetricFloat, 0.004, metric.UnitSeconds),
		ExpectMetric("maximum", "", metric.MetricFloat, 0.005, metric.UnitSeconds),
		ExpectMetric("minimum", "", metric.MetricFloat, 0.002, metric.UnitSeconds),
		ExpectMetric("available", "", metric.MetricFloat, 60.0, metric.UnitPercent),
		ExpectMetric("count", "", metric.MetricNumber, count, ""),
	}

	assert.Equal(t, 1, crs.Length())
	AssertMetrics(t, expected, crs.Get(0).Metrics)
	assert.False(t, crs.Available)
}
