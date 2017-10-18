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

// Package metric provides the messaging structures specific to metric reporting
package metric

import (
	"errors"
	"fmt"
	"strconv"
)

type Metric struct {
	Type       int         `json:"-"` // does not export to json
	TypeString string      `json:"type"`
	Dimension  string      `json:"dimension"`
	Name       string      `json:"name"`
	Unit       string      `json:"unit"`
	Value      interface{} `json:"value"`
}

func (m *Metric) String() string {
	return fmt.Sprintf("{type=%v, dimension=%v, name=%v, unit=%v, value=%v}",
		m.TypeString, m.Dimension, m.Name, m.Unit, m.Value)
}

func UnitToString(unit int) string {
	switch unit {
	case MetricString:
		return "string"
	case MetricBool:
		return "bool"
	case MetricNumber:
		return "int64"
	case MetricFloat:
		return "double"
	}
	return "undefined"
}

func NewMetric(name, metricDimension string, internalMetricType int, value interface{}, unit string) *Metric {
	if len(metricDimension) == 0 {
		metricDimension = "none"
	}
	metric := &Metric{
		Type:       internalMetricType,
		TypeString: UnitToString(internalMetricType),
		Dimension:  metricDimension,
		Name:       name,
		Value:      value,
		Unit:       unit,
	}
	return metric
}

func NewPercentMetricFromInt(name, metricDimension string, portion, total int) *Metric {
	if len(metricDimension) == 0 {
		metricDimension = "none"
	}
	metric := &Metric{
		Type:       MetricFloat,
		TypeString: UnitToString(MetricFloat),
		Dimension:  metricDimension,
		Name:       name,
		Value:      float64(100*portion) / float64(total),
		Unit:       UnitPercent,
	}
	return metric
}

func (m *Metric) ToString() (string, error) {
	if m.Type != MetricString {
		return "", errors.New("Invalid coercion to String")
	}
	if m.Value == nil {
		return "", nil
	}
	value, ok := m.Value.(string)
	if !ok {
		return "", errors.New("Invalid coercion to String")
	}
	return value, nil
}

func (m *Metric) ToUint64() (uint64, error) {
	if m.Type != MetricNumber {
		return 0, errors.New("Invalid coercion to Uint64")
	}
	value, ok := m.Value.(uint64)
	if !ok {
		return 0, errors.New("Invalid coercion to Uint64")
	}
	return value, nil
}

func (m *Metric) ToInt64() (int64, error) {
	if m.Type != MetricNumber {
		return 0, errors.New("Invalid coercion to int64")
	}
	value, ok := m.Value.(int64)
	if !ok {
		return 0, errors.New("Invalid coercion to int64")
	}
	return value, nil
}

func (m *Metric) ToInt32() (int32, error) {
	if m.Type != MetricNumber {
		return 0, errors.New("Invalid coercion to int32")
	}
	value, ok := m.Value.(int32)
	if !ok {
		return 0, errors.New("Invalid coercion to int32")
	}
	return value, nil
}

func (m *Metric) ToFloat64() (float64, error) {
	if m.Type == MetricString {
		strVal, ok := m.Value.(string)
		if !ok {
			return 0, errors.New("Wrong type for string metric")
		}
		return strToFloat(strVal)
	}
	value, ok := m.Value.(float64)
	if !ok {
		// Let's try harder...
		asStr := fmt.Sprintf("%v", m.Value)
		return strToFloat(asStr)
	}
	return value, nil
}

func strToFloat(strVal string) (float64, error) {
	asFloat, err := strconv.ParseFloat(strVal, 32)
	if err != nil {
		return 0, errors.New("Invalid coercion to float64")
	} else {
		return asFloat, nil
	}
}