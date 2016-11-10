package metric

import (
	"errors"
)

const (
	MetricString = iota
	MetricBool   = iota
	MetricNumber = iota
	MetricFloat  = iota
)

type Metric struct {
	Type       int         `json:"-"` // does not export to json
	TypeString string      `json:"type"`
	Dimension  string      `json:"dimension"`
	Name       string      `json:"name"`
	Unit       string      `json:"unit"`
	Value      interface{} `json:"value"`
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

func (m *Metric) ToFloat64() (float64, error) {
	if m.Type != MetricFloat {
		return 0, errors.New("Invalid coercion to float64")
	}
	value, ok := m.Value.(float64)
	if !ok {
		return 0, errors.New("Invalid coercion to float64")
	}
	return value, nil
}
