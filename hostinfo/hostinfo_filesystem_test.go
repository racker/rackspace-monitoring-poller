package hostinfo_test

import (
	"fmt"
	"testing"

	"github.com/racker/rackspace-monitoring-poller/hostinfo"
	protocol_hostinfo "github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/shirou/gopsutil/disk"
	"github.com/stretchr/testify/assert"
)

func TestNewHostInfoFilesystem(t *testing.T) {
	tests := []struct {
		name     string
		base     *protocol_hostinfo.HostInfoBase
		expected *hostinfo.HostInfoFilesystem
	}{
		{
			name: "Happy path",
			base: &protocol_hostinfo.HostInfoBase{
				Type: "test_type",
			},
			expected: &hostinfo.HostInfoFilesystem{
				HostInfoBase: protocol_hostinfo.HostInfoBase{
					Type: "test_type",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hostinfo.NewHostInfoFilesystem(tt.base)
			assert.Equal(
				t, got,
				tt.expected, fmt.Sprintf(
					"NewHostInfoFilesystem() = %v, expected %v",
					got, tt.expected))
		})
	}
}

func TestHostInfoFilesystem_Run(t *testing.T) {
	expectedPartitions, err := disk.Partitions(false)
	if err != nil {
		t.Skip("We cannot get filesystem info right now.  Skipping")
	}

	tests := []struct {
		name        string
		expected    map[string]*metric.Metric
		expectedErr bool
	}{
		{
			name: "Happy path",
			expected: map[string]*metric.Metric{
				"dir_name": &metric.Metric{
					Name:       "dir_name",
					Dimension:  "none",
					Type:       metric.MetricString,
					TypeString: "string",
					Value:      expectedPartitions[0].Mountpoint,
					Unit:       "",
				},
				"dev_name": &metric.Metric{
					Name:       "dev_name",
					Dimension:  "none",
					Type:       metric.MetricString,
					TypeString: "string",
					Value:      expectedPartitions[0].Device,
					Unit:       "",
				},
				"sys_type_name": &metric.Metric{
					Name:       "sys_type_name",
					Dimension:  "none",
					Type:       metric.MetricString,
					TypeString: "string",
					Value:      expectedPartitions[0].Fstype,
					Unit:       "",
				},
				"options": &metric.Metric{
					Name:       "options",
					Dimension:  "none",
					Type:       metric.MetricString,
					TypeString: "string",
					Value:      expectedPartitions[0].Opts,
					Unit:       "",
				},
			},
			expectedErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &hostinfo.HostInfoFilesystem{
				HostInfoBase: protocol_hostinfo.HostInfoBase{},
			}
			got, err := h.Run()
			if tt.expectedErr {
				assert.Error(t, err, fmt.Sprintf("HostInfoFilesystem.Run() error = %v, expectedErr %v", err, tt.expectedErr))
			} else {
				results := got.(*protocol_hostinfo.HostInfoFilesystemResult)
				assert.NotEmpty(t, results.Metrics[0].DirectoryName)
				assert.NotEmpty(t, results.Metrics[0].DeviceName)
				assert.NotEmpty(t, results.Metrics[0].Options)
				assert.NotZero(t, results.Metrics[0].Total)
				assert.NotZero(t, results.Metrics[0].Free)
				assert.NotZero(t, results.Metrics[0].Used)
				assert.NotZero(t, results.Metrics[0].Available)
			}
		})
	}
}
