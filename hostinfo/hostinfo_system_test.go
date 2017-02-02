package hostinfo_test

import (
	"fmt"
	"runtime"
	"testing"

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
				result := got.(*protocol_hostinfo.HostInfoSystemResult)
				assert.NotEmpty(t, result.Metrics.Name)
				assert.NotEmpty(t, result.Metrics.Arch)
				assert.NotEmpty(t, result.Metrics.Version)
				assert.NotEmpty(t, result.Metrics.VendorName)
				assert.NotEmpty(t, result.Metrics.Vendor)
				assert.NotEmpty(t, result.Metrics.VendorVersion)
			}
		})
	}
}
