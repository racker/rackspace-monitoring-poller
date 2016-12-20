package hostinfo_test

import (
	"reflect"
	"testing"

	"github.com/racker/rackspace-monitoring-poller/check"
	protocol_hostinfo "github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"

	"github.com/racker/rackspace-monitoring-poller/hostinfo"
	"github.com/shirou/gopsutil/mem"
)

func TestNewHostInfoMemory(t *testing.T) {
	tests := []struct {
		name     string
		base     *protocol_hostinfo.HostInfoBase
		expected *hostinfo.HostInfoMemory
	}{
		{
			name: "Happy path",
			base: &protocol_hostinfo.HostInfoBase{
				Type: "test_type",
			},
			expected: &hostinfo.HostInfoMemory{
				HostInfoBase: protocol_hostinfo.HostInfoBase{
					Type: "test_type",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hostinfo.NewHostInfoMemory(tt.base); !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NewHostInfoMemory() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestHostInfoMemory_Run(t *testing.T) {
	v, err := mem.VirtualMemory()
	if err != nil {
		t.Skip("We cannot get virtual memory info right now.  Skipping")
	}
	s, err := mem.SwapMemory()
	if err != nil {
		t.Skip("We cannot get swap memory info right now.  Skipping")
	}
	tests := []struct {
		name        string
		expected    map[string]*metric.Metric
		expectedErr bool
	}{
		{
			name: "Happy path",
			expected: map[string]*metric.Metric{
				"Total": &metric.Metric{
					Name:       "Total",
					Dimension:  "none",
					Type:       metric.MetricNumber,
					TypeString: "int64",
					Value:      v.Total,
					Unit:       "",
				},
				"SwapTotal": &metric.Metric{
					Name:       "SwapTotal",
					Dimension:  "none",
					Type:       metric.MetricNumber,
					TypeString: "int64",
					Value:      s.Total,
					Unit:       "",
				},
			},
			expectedErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &hostinfo.HostInfoMemory{
				HostInfoBase: protocol_hostinfo.HostInfoBase{},
			}
			got, err := h.Run()
			if (err != nil) != tt.expectedErr {
				t.Errorf("HostInfoMemory.Run() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}
			// loop through sample set and validate it exists in result
			for name, sample_metric := range tt.expected {
				// iterate through result
				var isFound, isValidated = false, false
				for _, got_check_result := range got.Metrics {
					if got_check_result.Metrics[name] != nil {
						isFound = true
						if reflect.DeepEqual(got_check_result.Metrics[name], sample_metric) {
							isValidated = true
						} else {
							t.Log("not equal ", sample_metric, got_check_result.Metrics[name])
						}
					}
				}
				if !isFound {
					t.Errorf("Metric not found in result = %v ", name)
				}
				if !isValidated {
					t.Errorf("Metric did not have correct values = %v ", sample_metric)
				}

			}
		})
	}
}

func TestHostInfoMemory_BuildResult(t *testing.T) {
	tests := []struct {
		name     string
		crs      *check.CheckResultSet
		expected *protocol_hostinfo.HostInfoMemoryResult
	}{
		{
			name: "Happy path",
			crs: &check.CheckResultSet{
				Metrics: []*check.CheckResult{
					&check.CheckResult{
						Metrics: map[string]*metric.Metric{
							"UsedPercentage": &metric.Metric{
								Name:       "UsedPercentage",
								Dimension:  "none",
								Type:       metric.MetricFloat,
								TypeString: "double",
								Value:      99.5,
								Unit:       "",
							},
							"Free": &metric.Metric{
								Name:       "Free",
								Dimension:  "none",
								Type:       metric.MetricNumber,
								TypeString: "int64",
								Value:      uint64(55),
								Unit:       "",
							},
							"Total": &metric.Metric{
								Name:       "Total",
								Dimension:  "none",
								Type:       metric.MetricNumber,
								TypeString: "int64",
								Value:      uint64(100000000),
								Unit:       "",
							},
							"Used": &metric.Metric{
								Name:       "Used",
								Dimension:  "none",
								Type:       metric.MetricNumber,
								TypeString: "int64",
								Value:      uint64(45),
								Unit:       "",
							},
							"SwapUsedPercentage": &metric.Metric{
								Name:       "SwapUsedPercentage",
								Dimension:  "none",
								Type:       metric.MetricFloat,
								TypeString: "double",
								Value:      99.5,
								Unit:       "",
							},
							"SwapFree": &metric.Metric{
								Name:       "SwapFree",
								Dimension:  "none",
								Type:       metric.MetricNumber,
								TypeString: "int64",
								Value:      uint64(55),
								Unit:       "",
							},
							"SwapTotal": &metric.Metric{
								Name:       "SwapTotal",
								Dimension:  "none",
								Type:       metric.MetricNumber,
								TypeString: "int64",
								Value:      uint64(100000000),
								Unit:       "",
							},
							"SwapUsed": &metric.Metric{
								Name:       "SwapUsed",
								Dimension:  "none",
								Type:       metric.MetricNumber,
								TypeString: "int64",
								Value:      uint64(45),
								Unit:       "",
							},
						},
					},
				},
			},
			expected: &protocol_hostinfo.HostInfoMemoryResult{
				Metrics: protocol_hostinfo.HostInfoMemoryMetrics{
					SwapFree:           55,
					SwapTotal:          100000000,
					SwapUsed:           45,
					SwapUsedPercentage: 99.5,
					Free:               55,
					ActualFree:         55,
					Total:              100000000,
					Used:               45,
					ActualUsed:         45,
					UsedPercentage:     99.5,
					RAM:                95,
				},
				Timestamp: utils.NowTimestampMillis(),
			},
		},
		{
			name: "Crs is nil",
			crs:  nil,
			expected: &protocol_hostinfo.HostInfoMemoryResult{
				Timestamp: utils.NowTimestampMillis(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &hostinfo.HostInfoMemory{
				HostInfoBase: protocol_hostinfo.HostInfoBase{},
			}
			if got := h.BuildResult(tt.crs); !reflect.DeepEqual(got.(*protocol_hostinfo.HostInfoMemoryResult).Metrics, tt.expected.Metrics) {
				t.Errorf("HostInfoMemory.BuildResult() = %v, expected %v", got.(*protocol_hostinfo.HostInfoMemoryResult).Metrics, tt.expected.Metrics)
			}
		})
	}
}
