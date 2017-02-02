package hostinfo_test

import (
	"fmt"
	"testing"

	"github.com/racker/rackspace-monitoring-poller/hostinfo"
	protocol_hostinfo "github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/shirou/gopsutil/process"
	"github.com/stretchr/testify/assert"
)

func getTimeMetric(metric_type string, pr *process.Process, t *testing.T) map[string]*metric.Metric {
	times, err := pr.Times()
	if err != nil {
		t.Error("Failed to get process times")
	}
	var value float64 = 0
	switch metric_type {
	case "time_user":
		value = times.User
		break
	case "time_sys":
		value = times.System
		break
	case "time_total":
		value = times.Total()
		break
	}
	return map[string]*metric.Metric{
		metric_type: &metric.Metric{
			Name:       metric_type,
			Dimension:  "none",
			Type:       metric.MetricFloat,
			TypeString: "double",
			Value:      value,
			Unit:       "",
		},
	}

}

func TestNewHostInfoProcesses(t *testing.T) {
	tests := []struct {
		name     string
		base     *protocol_hostinfo.HostInfoBase
		expected *hostinfo.HostInfoProcesses
	}{
		{
			name: "Happy path",
			base: &protocol_hostinfo.HostInfoBase{
				Type: "test_type",
			},
			expected: &hostinfo.HostInfoProcesses{
				HostInfoBase: protocol_hostinfo.HostInfoBase{
					Type: "test_type",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hostinfo.NewHostInfoProcesses(tt.base)
			assert.Equal(
				t, got,
				tt.expected, fmt.Sprintf(
					"NewHostInfoProcesses() = %v, expected %v",
					got, tt.expected))
		})
	}
}

func TestHostInfoProcesses_Run(t *testing.T) {
	expectedPids, err := process.Pids()
	if err != nil {
		t.Skip("Unable to get host process information.  Skipping for now", err)
	}
	expectedProcess, err := process.NewProcess(expectedPids[0])
	if err != nil {
		t.Skip("Unable to get the process pid. Skipping!", err)
	}
	expectedCwd, err := expectedProcess.Cwd()
	if err != nil {
		t.Skip("CWD currently not implemented.  Skipping", err)
	}
	tests := []struct {
		name        string
		expected    map[string]*metric.Metric
		expectedErr bool
	}{
		{
			name: "Happy path",
			expected: map[string]*metric.Metric{
				"pid": &metric.Metric{
					Name:       "pid",
					Dimension:  "none",
					Type:       metric.MetricNumber,
					TypeString: "int32",
					Value:      expectedProcess.Pid,
					Unit:       "",
				},
				"state_name": &metric.Metric{
					Name:       "state_name",
					Dimension:  "none",
					Type:       metric.MetricString,
					TypeString: "string",
					Value: func() string {
						name, err := expectedProcess.Name()
						if err != nil {
							t.Error("Unable to get the process name")
						}
						return name
					},
					Unit: "",
				},
				"exe_cwd": &metric.Metric{
					Name:       "exe_cwd",
					Dimension:  "none",
					Type:       metric.MetricString,
					TypeString: "string",
					Value:      expectedCwd,
					Unit:       "",
				},
				"exe_root": &metric.Metric{
					Name:       "exe_cwd",
					Dimension:  "none",
					Type:       metric.MetricString,
					TypeString: "string",
					Value: func() string {
						root, err := expectedProcess.Exe()
						if err != nil {
							t.Error("Unable to get the process root executable")
						}
						return root
					},
					Unit: "",
				},
				"time_start_time": &metric.Metric{
					Name:       "time_start_time",
					Dimension:  "none",
					Type:       metric.MetricNumber,
					TypeString: "int64",
					Value: func() int64 {
						createTime, err := expectedProcess.CreateTime()
						if err != nil {
							t.Error("Unable to get the process created time")
						}
						return createTime
					},
					Unit: "",
				},
			},
			expectedErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &hostinfo.HostInfoProcesses{
				HostInfoBase: protocol_hostinfo.HostInfoBase{},
			}
			got, _ := h.Run()
			if tt.expectedErr {
				assert.Error(t, err, fmt.Sprintf("HostInfoProcesses.Run() error = %v, expectedErr %v", err, tt.expectedErr))
			} else {
				result := got.(*protocol_hostinfo.HostInfoProcessesResult)
				assert.NotZero(t, len(result.Metrics))
			}
		})
	}
}
