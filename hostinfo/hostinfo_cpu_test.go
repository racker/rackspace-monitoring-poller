package hostinfo_test

import (
	"testing"

	"fmt"

	"github.com/racker/rackspace-monitoring-poller/hostinfo"
	protocol_hostinfo "github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
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
				result := got.(*protocol_hostinfo.HostInfoCpuResult)
				assert.NotEmpty(t, result.Metrics[0].Name)
				assert.NotEmpty(t, result.Metrics[0].Model)
				assert.NotEmpty(t, result.Metrics[0].Vendor)
				assert.NotZero(t, result.Metrics[0].Mhz)
				assert.NotZero(t, result.Metrics[0].Total)
				assert.NotZero(t, result.Metrics[0].TotalCores)
				assert.NotZero(t, result.Metrics[0].TotalSockets)
				assert.NotZero(t, result.Metrics[0].User)
			}
		})
	}
}
