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

	"fmt"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBase_GetTargetIP(t *testing.T) {

	cb := &check.Base{}

	cb.IpAddresses = map[string]string{
		"host1": "192.168.0.1",
	}
	targetAlias := "host1"
	cb.TargetAlias = &targetAlias

	result, err := cb.GetTargetIP()
	assert.NoError(t, err)
	assert.Equal(t, "192.168.0.1", result)
}

func TestBase_GetTargetIP_mismatch(t *testing.T) {

	cb := &check.Base{}

	cb.IpAddresses = map[string]string{
		"host1": "192.168.0.1",
	}
	targetAlias := "totallyWrongHost"
	cb.TargetAlias = &targetAlias

	_, err := cb.GetTargetIP()
	assert.Error(t, err)
}

func TestBase_GetWaitPeriod(t *testing.T) {
	cb := &check.Base{}

	cb.Period = 5

	assert.Equal(t, 5*time.Second, cb.GetWaitPeriod())
}

func TestBase_Cancel(t *testing.T) {

	root := context.Background()

	ch, err := check.NewCheck(root, json.RawMessage(`{
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
	  }`))
	require.NoError(t, err)
	require.NotNil(t, ch)

	// I know, looks weird, but pre-cancel it since channels are cool like that
	ch.Cancel()
	// and this timebox should finish immediately
	completed := utils.Timebox(t, 100*time.Millisecond, func(t *testing.T) {
		<-ch.Done()
	})

	assert.True(t, completed, "cancellation channel never notified")
}

func TestBase_IsDisabled(t *testing.T) {
	content := `{
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
	  "disabled":%s
	  }`

	tests := []struct {
		name     string
		disabled string
		expected bool
	}{
		{
			name:     "enabled",
			disabled: "false",
			expected: false,
		},
		{
			name:     "disabled",
			disabled: "true",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch, err := check.NewCheck(context.Background(), json.RawMessage(fmt.Sprintf(content, tt.disabled)))
			require.NoError(t, err)

			assert.Equal(t, tt.expected, ch.IsDisabled())
		})
	}
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
