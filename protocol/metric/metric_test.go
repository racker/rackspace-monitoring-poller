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

package metric_test

import (
	"testing"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/stretchr/testify/assert"
)

func TestMetric_ToFloat64(t *testing.T) {
	m := metric.NewMetric("someFloat", "", metric.MetricFloat, 6.0, "")
	floatVal, err := m.ToFloat64()
	assert.NoError(t, err)
	assert.Equal(t, 6.0, floatVal)
}

func TestMetric_ToFloat64_FromNumber(t *testing.T) {
	m := metric.NewMetric("someInt", "", metric.MetricNumber, 5, "")

	floatVal, err := m.ToFloat64()
	assert.NoError(t, err)

	assert.Equal(t, 5.0, floatVal)
}

func TestMetric_ToFloat64_FromString(t *testing.T) {
	m := metric.NewMetric("someStringWithFloat", "", metric.MetricString, "7.0", "")

	floatVal, err := m.ToFloat64()
	assert.NoError(t, err)

	assert.Equal(t, 7.0, floatVal)
}

func TestMetric_ToFloat64_FromStringNotString(t *testing.T) {
	m := metric.NewMetric("someStringWithFloat", "", metric.MetricString, 7.0, "")

	_, err := m.ToFloat64()
	assert.EqualError(t, err, "Wrong type for string metric")
}

func TestMetric_ToFloat64_FromStringNotNumber(t *testing.T) {
	m := metric.NewMetric("someStringWithStatus", "", metric.MetricString, "OK", "")

	_, err := m.ToFloat64()
	assert.EqualError(t, err, "Invalid coercion to float64")
}
