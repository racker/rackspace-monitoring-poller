package hostinfo_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/hostinfo"
	protocol_hostinfo "github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/shirou/gopsutil/host"
	"github.com/stretchr/testify/assert"
)

func TestNewHostInfoSystem(t *testing.T) {
	tests := []struct {
		name     string
		base     *protocol_hostinfo.HostInfoBase
		expected *hostinfo.HostInfoSystem
	}{
		{
			name: "Happy path",
			base: &protocol_hostinfo.HostInfoBase{
				Type: "test_type",
			},
			expected: &hostinfo.HostInfoSystem{
				HostInfoBase: protocol_hostinfo.HostInfoBase{
					Type: "test_type",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hostinfo.NewHostInfoSystem(tt.base)
			assert.Equal(
				t, got,
				tt.expected, fmt.Sprintf(
					"NewHostInfoSystem() = %v, expected %v",
					got, tt.expected))
		})
	}
}

func TestHostInfoSystem_Run(t *testing.T) {
	expectedInfo, err := host.Info()
	if err != nil {
		t.Skip("Host info is unavailable. Skipping", err)
	}
	tests := []struct {
		name        string
		expected    map[string]*metric.Metric
		expectedErr bool
	}{
		{
			name: "Happy path",
			expected: map[string]*metric.Metric{
				"arch": &metric.Metric{
					Name:       "arch",
					Dimension:  "none",
					Type:       metric.MetricString,
					TypeString: "string",
					Value:      runtime.GOOS,
					Unit:       "",
				},
				"name": &metric.Metric{
					Name:       "name",
					Dimension:  "none",
					Type:       metric.MetricString,
					TypeString: "string",
					Value:      expectedInfo.OS,
					Unit:       "",
				},
				"version": &metric.Metric{
					Name:       "version",
					Dimension:  "none",
					Type:       metric.MetricString,
					TypeString: "string",
					Value:      expectedInfo.KernelVersion,
					Unit:       "",
				},
				"vendor_name": &metric.Metric{
					Name:       "vendor_name",
					Dimension:  "none",
					Type:       metric.MetricString,
					TypeString: "string",
					Value:      expectedInfo.OS,
					Unit:       "",
				},
				"vendor": &metric.Metric{
					Name:       "vendor",
					Dimension:  "none",
					Type:       metric.MetricString,
					TypeString: "string",
					Value:      expectedInfo.Platform,
					Unit:       "",
				},
				"vendor_version": &metric.Metric{
					Name:       "vendor_version",
					Dimension:  "none",
					Type:       metric.MetricString,
					TypeString: "string",
					Value:      expectedInfo.PlatformVersion,
					Unit:       "",
				},
			},
			expectedErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &hostinfo.HostInfoSystem{
				HostInfoBase: protocol_hostinfo.HostInfoBase{},
			}
			got, err := h.Run()
			if tt.expectedErr {
				assert.Error(t, err, fmt.Sprintf("HostInfoSystem.Run() error = %v, expectedErr %v", err, tt.expectedErr))
			} else {
				// loop through sample set and validate it exists in result
				for name, sample_metric := range tt.expected {
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

func TestHostInfoSystem_BuildResult(t *testing.T) {
	tests := []struct {
		name     string
		crs      *check.ResultSet
		expected *protocol_hostinfo.HostInfoSystemResult
	}{
		{
			name: "Happy path",
			crs: &check.ResultSet{
				Metrics: []*check.Result{
					&check.Result{
						Metrics: map[string]*metric.Metric{
							"arch": &metric.Metric{
								Name:       "arch",
								Dimension:  "none",
								Type:       metric.MetricString,
								TypeString: "string",
								Value:      "runtime.GOOS",
								Unit:       "",
							},
							"name": &metric.Metric{
								Name:       "name",
								Dimension:  "none",
								Type:       metric.MetricString,
								TypeString: "string",
								Value:      "info.OS",
								Unit:       "",
							},
							"version": &metric.Metric{
								Name:       "version",
								Dimension:  "none",
								Type:       metric.MetricString,
								TypeString: "string",
								Value:      "info.KernelVersion",
								Unit:       "",
							},
							"vendor_name": &metric.Metric{
								Name:       "vendor_name",
								Dimension:  "none",
								Type:       metric.MetricString,
								TypeString: "string",
								Value:      "info.OS",
								Unit:       "",
							},
							"vendor": &metric.Metric{
								Name:       "vendor",
								Dimension:  "none",
								Type:       metric.MetricString,
								TypeString: "string",
								Value:      "info.Platform",
								Unit:       "",
							},
							"vendor_version": &metric.Metric{
								Name:       "vendor_version",
								Dimension:  "none",
								Type:       metric.MetricString,
								TypeString: "string",
								Value:      "info.PlatformVersion",
								Unit:       "",
							},
						},
					},
				},
			},
			expected: &protocol_hostinfo.HostInfoSystemResult{
				Metrics: protocol_hostinfo.HostInfoSystemMetrics{
					Vendor:        "info.Platform",
					Arch:          "runtime.GOOS",
					Name:          "info.OS",
					VendorName:    "info.OS",
					VendorVersion: "info.PlatformVersion",
					Version:       "info.KernelVersion",
				},
			},
		},
		{
			name:     "Crs is nil",
			crs:      nil,
			expected: &protocol_hostinfo.HostInfoSystemResult{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &hostinfo.HostInfoSystem{
				HostInfoBase: protocol_hostinfo.HostInfoBase{},
			}
			got := h.BuildResult(tt.crs)
			assert.Equal(
				t, tt.expected.Metrics,
				got.(*protocol_hostinfo.HostInfoSystemResult).Metrics,
				fmt.Sprintf("HostInfoSystem.BuildResult() = %v, expected %v", got.(*protocol_hostinfo.HostInfoSystemResult).Metrics, tt.expected.Metrics))

		})
	}
}
