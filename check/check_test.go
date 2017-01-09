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
	"encoding/json"
	"testing"
	"time"

	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestCheckBase_Cancel(t *testing.T) {

	root := context.Background()
	cancelCtx, cancelFunc := context.WithCancel(root)

	ch := check.NewCheck(json.RawMessage(`{
	  "id":"chPzATCP",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"port":0,"ssl":false},
	  "type":"remote.tcp",
	  "timeout":1,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"",
	  "disabled":true
	  }`), cancelCtx, cancelFunc)
	require.NotNil(t, ch)

	// I know, looks weird, but pre-cancel it since channels are cool like that
	ch.Cancel()
	// and this timebox should finish immediately
	completed := Timebox(t, 100*time.Millisecond, func(t *testing.T) {
		<-ch.Done()
	})

	assert.True(t, completed, "cancellation channel never notified")
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

// Timebox is used for putting a time bounds around a chunk of code, given as the function boxed.
// NOTE that if the duration d elapses, then boxed will be left to run off in its go-routine...it can't be
// forcefully terminated.
// This function can be used outside of a unit test context by passing nil for t
// Returns true if boxed finished before duration d elapsed.
func Timebox(t *testing.T, d time.Duration, boxed func(t *testing.T)) bool {
	timer := time.NewTimer(d)
	completed := make(chan struct{})

	go func() {
		boxed(t)
		close(completed)
	}()

	select {
	case <-timer.C:
		if t != nil {
			t.Fatal("Timebox expired")
		}
		return false
	case <-completed:
		timer.Stop()
		return true
	}
}

func TestTimebox_Quick(t *testing.T) {
	result := Timebox(t, 1*time.Second, func(t *testing.T) {
		time.Sleep(1 * time.Millisecond)
	})
	assert.True(t, result)
}

func TestTimebox_TimesOut(t *testing.T) {
	result := Timebox(nil, 1*time.Millisecond, func(t *testing.T) {
		time.Sleep(100 * time.Millisecond)
	})

	assert.False(t, result)
}
