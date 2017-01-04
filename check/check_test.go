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
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCheckBase_GetTargetIP(t *testing.T) {

	cb := &check.CheckBase{}

	cb.IpAddresses = map[string]string{
		"host1": "192.168.0.1",
	}
	targetAlias := "host1"
	cb.TargetAlias = &targetAlias

	result, err := cb.GetTargetIP()
	assert.NoError(t, err)
	assert.Equal(t, "192.168.0.1", result)
}

func TestCheckBase_GetTargetIP_mismatch(t *testing.T) {

	cb := &check.CheckBase{}

	cb.IpAddresses = map[string]string{
		"host1": "192.168.0.1",
	}
	targetAlias := "totallyWrongHost"
	cb.TargetAlias = &targetAlias

	_, err := cb.GetTargetIP()
	assert.Error(t, err)
}

func TestCheckBase_GetWaitPeriod(t *testing.T) {
	cb := &check.CheckBase{}

	cb.Period = 5

	assert.Equal(t, 5*time.Second, cb.GetWaitPeriod())
}

type ExpectedMetric struct {
	metric.Metric

	ExpectNonZeroValue bool
	IgnoreValue        bool
}

func ExpectMetric(name, metricDimension string, internalMetricType int, value interface{}, unit string) *ExpectedMetric {
	return &ExpectedMetric{
		Metric: *metric.NewMetric(name, metricDimension, internalMetricType, value, unit),
	}
}

func (m *ExpectedMetric) ButNonZeroValue() *ExpectedMetric {
	m.ExpectNonZeroValue = true
	return m
}

func (m *ExpectedMetric) ButIgnoreValue() *ExpectedMetric {
	m.IgnoreValue = true
	return m
}

func AssertMetrics(t *testing.T, expected []*ExpectedMetric, actual map[string]*metric.Metric) {

	require.Len(t, actual, len(expected))

	for _, m := range expected {
		actualM, ok := actual[m.Name]
		require.True(t, ok, "existence of %s", m.Name)

		assert.Equal(t, m.Type, actualM.Type)
		if m.ExpectNonZeroValue {
			assert.NotZero(t, actualM.Value, "value of %s", m.Name)
		} else if !m.IgnoreValue {
			assert.Equal(t, m.Value, actualM.Value, "value of %s", m.Name)
		}
		assert.Equal(t, m.Unit, actualM.Unit, "unit of %s", m.Name)

	}
}

func Timebox(t *testing.T, d time.Duration, boxed func(t *testing.T)) {
	timer := time.NewTimer(d)
	completed := make(chan int)

	go func() {
		boxed(t)
		completed <- 1
	}()

	select {
	case <-timer.C:
		t.Fatal("Timebox expired")
	case <-completed:
		timer.Stop()
	}
}
