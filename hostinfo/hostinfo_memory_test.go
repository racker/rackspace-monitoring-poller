package hostinfo_test

import (
	"fmt"
	"testing"

	"github.com/racker/rackspace-monitoring-poller/hostinfo"
	protocol_hostinfo "github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/shirou/gopsutil/mem"
	"github.com/stretchr/testify/assert"
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
			got := hostinfo.NewHostInfoMemory(tt.base)
			assert.Equal(
				t, got,
				tt.expected, fmt.Sprintf(
					"NewHostInfoMemory() = %v, expected %v",
					got, tt.expected))
		})
	}
}

func TestHostInfoMemory_Run(t *testing.T) {
	expectedVirtualMem, err := mem.VirtualMemory()
	if err != nil {
		t.Skip("We cannot get virtual memory info right now.  Skipping")
	}
	expectedSwapMem, err := mem.SwapMemory()
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
					Value:      expectedVirtualMem.Total,
					Unit:       "",
				},
				"SwapTotal": &metric.Metric{
					Name:       "SwapTotal",
					Dimension:  "none",
					Type:       metric.MetricNumber,
					TypeString: "int64",
					Value:      expectedSwapMem.Total,
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
			got, _ := h.Run()
			if tt.expectedErr {
				assert.Error(t, err, fmt.Sprintf("HostInfoMemory.Run() error = %v, expectedErr %v", err, tt.expectedErr))
			} else {
				result := got.(*protocol_hostinfo.HostInfoMemoryResult)
				assert.NotZero(t, result.Metrics.ActualFree)
				assert.NotZero(t, result.Metrics.Free)
				assert.NotZero(t, result.Metrics.Used)
				assert.NotZero(t, result.Metrics.ActualUsed)
				assert.NotZero(t, result.Metrics.Total)
				assert.NotZero(t, result.Metrics.RAM)
			}
		})
	}
}
