package hostinfo_test

import (
	"testing"

	"fmt"

	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/hostinfo"
	protocol_hostinfo "github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/shirou/gopsutil/cpu"
	"github.com/stretchr/testify/assert"
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
			assert.Equal(
				t, got,
				tt.expected, fmt.Sprintf(
					"NewHostInfoCpu() = %v, expected %v",
					got, tt.expected))
		})
	}
}

func TestHostInfoCpu_Run(t *testing.T) {
	expectedInfo, err := cpu.Info()
	if err != nil {
		t.Skip("We cannot get cpu info right now.  Skipping")
	}
	expectedCoreCount, _ := cpu.Counts(true)

	tests := []struct {
		name                    string
		expectedResultSetLength int
		expectedSampleMetricSet map[string]*metric.Metric
		expectedErr             bool
	}{
		{
			name: "Happy path",
			expectedResultSetLength: len(expectedInfo),
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
					Value:      expectedInfo[0].ModelName,
					Unit:       "",
				},
				"total_cores": &metric.Metric{
					Name:       "total_cores",
					Dimension:  "none",
					Type:       metric.MetricNumber,
					TypeString: "int64",
					Value:      expectedCoreCount,
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
			if tt.expectedErr {
				assert.Error(t, err, fmt.Sprintf("HostInfoCpu.Run() error = %v, expectedErr %v", err, tt.expectedErr))
			} else {
				for name, sample_metric := range tt.expectedSampleMetricSet {
					// first, get the metric
					gotMetric := getMetricResults(name, got.Metrics)
					if gotMetric == nil {
						t.Errorf("Metric not found in result = %v ", name)
					} else {
						assert.Equal(
							t, sample_metric, gotMetric,
							fmt.Sprintf("Metric did not have correct values = %v ", sample_metric))
					}

				}
			}
		})
	}
}

func TestHostInfoCpu_BuildResult(t *testing.T) {
	tests := []struct {
		name     string
		crs      *check.ResultSet
		expected *protocol_hostinfo.HostInfoCpuResult
	}{
		{
			name: "Happy path",
			crs: &check.ResultSet{
				Metrics: []*check.Result{
					&check.Result{
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

			got := h.BuildResult(tt.crs)
			assert.Equal(
				t, tt.expected.Metrics,
				got.(*protocol_hostinfo.HostInfoCpuResult).Metrics,
				fmt.Sprintf("HostInfoCpu.BuildResult() = %v, expected %v", got.(*protocol_hostinfo.HostInfoCpuResult).Metrics, tt.expected.Metrics))

		})
	}
}
