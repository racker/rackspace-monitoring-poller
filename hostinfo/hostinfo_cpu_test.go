package hostinfo_test

import (
	"reflect"
	"testing"

	"fmt"

	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/hostinfo"
	protocol_hostinfo "github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/shirou/gopsutil/cpu"
)

func TestNewHostInfoCpu(t *testing.T) {
	tests := []struct {
		name     string
		base     *protocol_hostinfo.HostInfoBase
		expected *hostinfo.HostInfoCpu
	}{
		{
			name: "Happy path",
			base: &protocol_hostinfo.HostInfoBase{
				Type: "test_type",
			},
			expected: &hostinfo.HostInfoCpu{
				HostInfoBase: protocol_hostinfo.HostInfoBase{
					Type: "test_type",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hostinfo.NewHostInfoCpu(tt.base)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NewHostInfoCpu() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestHostInfoCpu_Run(t *testing.T) {
	info, err := cpu.Info()
	if err != nil {
		t.Skip("We cannot get cpu info right now.  Skipping")
	}
	stats, err := cpu.Times(true)
	if err != nil {
		t.Skip("We cannot get cpu stats right now.  Skipping")
	}
	coreCount, _ := cpu.Counts(true)

	tests := []struct {
		name                    string
		expectedResultSetLength int
		expectedSampleMetricSet map[string]*metric.Metric
		expectedErr             bool
	}{
		{
			name: "Happy path",
			expectedResultSetLength: len(info),
			expectedSampleMetricSet: map[string]*metric.Metric{
				"name": &metric.Metric{
					Name:       "name",
					Dimension:  "none",
					Type:       metric.MetricString,
					TypeString: "string",
					Value:      fmt.Sprintf("cpu.%d", 0),
					Unit:       "",
				},
				"model": &metric.Metric{
					Name:       "model",
					Dimension:  "none",
					Type:       metric.MetricString,
					TypeString: "string",
					Value:      info[0].ModelName,
					Unit:       "",
				},
				"total_cores": &metric.Metric{
					Name:       "total_cores",
					Dimension:  "none",
					Type:       metric.MetricNumber,
					TypeString: "int64",
					Value:      coreCount,
					Unit:       "",
				},
				"total": &metric.Metric{
					Name:       "total",
					Dimension:  "none",
					Type:       metric.MetricFloat,
					TypeString: "double",
					Value:      stats[0].Total(),
					Unit:       "",
				},
			},
			expectedErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &hostinfo.HostInfoCpu{
				HostInfoBase: protocol_hostinfo.HostInfoBase{},
			}
			got, err := h.Run()
			if (err != nil) != tt.expectedErr {
				t.Errorf("HostInfoCpu.Run() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}
			// loop through sample set and validate it exists in result
			for name, sample_metric := range tt.expectedSampleMetricSet {
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

func TestHostInfoCpu_BuildResult(t *testing.T) {
	tests := []struct {
		name     string
		crs      *check.CheckResultSet
		expected *protocol_hostinfo.HostInfoCpuResult
	}{
		{
			name: "Happy path",
			crs: &check.CheckResultSet{
				Metrics: []*check.CheckResult{
					&check.CheckResult{
						Metrics: map[string]*metric.Metric{
							"name": &metric.Metric{
								Name:       "name",
								Dimension:  "none",
								Type:       metric.MetricString,
								TypeString: "string",
								Value:      fmt.Sprintf("cpu.%d", 0),
								Unit:       "",
							},
							"model": &metric.Metric{
								Name:       "model",
								Dimension:  "none",
								Type:       metric.MetricString,
								TypeString: "string",
								Value:      "model test",
								Unit:       "",
							},
							"vendor": &metric.Metric{
								Name:       "vendor",
								Dimension:  "none",
								Type:       metric.MetricString,
								TypeString: "string",
								Value:      "vendor",
								Unit:       "",
							},
							"idle": &metric.Metric{
								Name:       "idle",
								Dimension:  "none",
								Type:       metric.MetricFloat,
								TypeString: "double",
								Value:      12.5,
								Unit:       "",
							},
							"irq": &metric.Metric{
								Name:       "irq",
								Dimension:  "none",
								Type:       metric.MetricFloat,
								TypeString: "double",
								Value:      11.5,
								Unit:       "",
							},
							"mhz": &metric.Metric{
								Name:       "mhz",
								Dimension:  "none",
								Type:       metric.MetricFloat,
								TypeString: "double",
								Value:      10.5,
								Unit:       "",
							},
							"nice": &metric.Metric{
								Name:       "nice",
								Dimension:  "none",
								Type:       metric.MetricFloat,
								TypeString: "double",
								Value:      9.5,
								Unit:       "",
							},
							"soft_irq": &metric.Metric{
								Name:       "soft_irq",
								Dimension:  "none",
								Type:       metric.MetricFloat,
								TypeString: "double",
								Value:      8.5,
								Unit:       "",
							},
							"stolen": &metric.Metric{
								Name:       "stolen",
								Dimension:  "none",
								Type:       metric.MetricFloat,
								TypeString: "double",
								Value:      8.6,
								Unit:       "",
							},
							"sys": &metric.Metric{
								Name:       "sys",
								Dimension:  "none",
								Type:       metric.MetricFloat,
								TypeString: "double",
								Value:      7.5,
								Unit:       "",
							},
							"total_cores": &metric.Metric{
								Name:       "total_cores",
								Dimension:  "none",
								Type:       metric.MetricNumber,
								TypeString: "int64",
								Value:      uint64(8),
								Unit:       "",
							},
							"total": &metric.Metric{
								Name:       "total",
								Dimension:  "none",
								Type:       metric.MetricFloat,
								TypeString: "double",
								Value:      10.5,
								Unit:       "",
							},
							"total_sockets": &metric.Metric{
								Name:       "total_sockets",
								Dimension:  "none",
								Type:       metric.MetricNumber,
								TypeString: "int64",
								Value:      uint64(8),
								Unit:       "",
							},
							"user": &metric.Metric{
								Name:       "user",
								Dimension:  "none",
								Type:       metric.MetricFloat,
								TypeString: "double",
								Value:      5.5,
								Unit:       "",
							},
							"wait": &metric.Metric{
								Name:       "wait",
								Dimension:  "none",
								Type:       metric.MetricFloat,
								TypeString: "double",
								Value:      4.5,
								Unit:       "",
							},
						},
					},
				},
			},
			expected: &protocol_hostinfo.HostInfoCpuResult{
				Metrics: []protocol_hostinfo.HostInfoCpuMetrics{
					protocol_hostinfo.HostInfoCpuMetrics{
						Idle:         12,
						Irq:          11.5,
						Mhz:          10.5,
						Model:        "model test",
						Name:         "cpu.0",
						Nice:         9.5,
						SoftIrq:      8.5,
						Stolen:       8.6,
						Sys:          7.5,
						Total:        10,
						TotalCores:   8,
						TotalSockets: 8,
						User:         5.5,
						Wait:         4.5,
						Vendor:       "vendor",
					},
				},
				Timestamp: utils.NowTimestampMillis(),
			},
		},
		{
			name: "Crs is nil",
			crs:  nil,
			expected: &protocol_hostinfo.HostInfoCpuResult{
				Timestamp: utils.NowTimestampMillis(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &hostinfo.HostInfoCpu{
				HostInfoBase: protocol_hostinfo.HostInfoBase{},
			}

			if got := h.BuildResult(tt.crs); !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("HostInfoCpu.BuildResult() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
