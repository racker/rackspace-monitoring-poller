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
				fmt.Println(got)
				// loop through sample set and validate it exists in result
				//for name, sample_metric := range tt.expected {
				// first, get the metric
				//gotMetric := getMetricResults(name, got.Metrics)
				//if gotMetric == nil {
				//	t.Errorf("Metric not found in result = %v ", name)
				//} else {
				//	assert.Equal(
				//		t, sample_metric, gotMetric,
				//		fmt.Sprintf("Metric did not have correct values = %v ", sample_metric))
				//}
				//}
			}
		})
	}
}
