package hostinfo_test

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/hostinfo"
	protocol_hostinfo "github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/shirou/gopsutil/host"
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
			if got := hostinfo.NewHostInfoSystem(tt.base); !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NewHostInfoSystem() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestHostInfoSystem_Run(t *testing.T) {
	info, err := host.Info()
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
					Value:      info.OS,
					Unit:       "",
				},
				"version": &metric.Metric{
					Name:       "version",
					Dimension:  "none",
					Type:       metric.MetricString,
					TypeString: "string",
					Value:      info.KernelVersion,
					Unit:       "",
				},
				"vendor_name": &metric.Metric{
					Name:       "vendor_name",
					Dimension:  "none",
					Type:       metric.MetricString,
					TypeString: "string",
					Value:      info.OS,
					Unit:       "",
				},
				"vendor": &metric.Metric{
					Name:       "vendor",
					Dimension:  "none",
					Type:       metric.MetricString,
					TypeString: "string",
					Value:      info.Platform,
					Unit:       "",
				},
				"vendor_version": &metric.Metric{
					Name:       "vendor_version",
					Dimension:  "none",
					Type:       metric.MetricString,
					TypeString: "string",
					Value:      info.PlatformVersion,
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
			if (err != nil) != tt.expectedErr {
				t.Errorf("HostInfoSystem.Run() error = %v, expectedErr %v", err, tt.expectedErr)
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

func TestHostInfoSystem_BuildResult(t *testing.T) {
	tests := []struct {
		name     string
		crs      *check.CheckResultSet
		expected *protocol_hostinfo.HostInfoSystemResult
	}{
		{
			name: "Happy path",
			crs: &check.CheckResultSet{
				Metrics: []*check.CheckResult{
					&check.CheckResult{
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
			if got := h.BuildResult(tt.crs); !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("HostInfoSystem.BuildResult() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
