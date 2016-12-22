package hostinfo_test

import (
	"fmt"
	"testing"

	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/hostinfo"
	protocol_hostinfo "github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"
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

func TestHostInfoFilesystem_BuildResult(t *testing.T) {
	tests := []struct {
		name     string
		crs      *check.CheckResultSet
		expected *protocol_hostinfo.HostInfoFilesystemResult
	}{
		{
			name: "Happy path",
			crs: &check.CheckResultSet{
				Metrics: []*check.CheckResult{
					&check.CheckResult{
						Metrics: map[string]*metric.Metric{
							"dir_name": &metric.Metric{
								Name:       "dir_name",
								Dimension:  "none",
								Type:       metric.MetricString,
								TypeString: "string",
								Value:      "mount point",
								Unit:       "",
							},
							"dev_name": &metric.Metric{
								Name:       "dev_name",
								Dimension:  "none",
								Type:       metric.MetricString,
								TypeString: "string",
								Value:      "partition device",
								Unit:       "",
							},
							"sys_type_name": &metric.Metric{
								Name:       "sys_type_name",
								Dimension:  "none",
								Type:       metric.MetricString,
								TypeString: "string",
								Value:      "partition fstype",
								Unit:       "",
							},
							"options": &metric.Metric{
								Name:       "options",
								Dimension:  "none",
								Type:       metric.MetricString,
								TypeString: "string",
								Value:      "partition options",
								Unit:       "",
							},
							"total": &metric.Metric{
								Name:       "total",
								Dimension:  "none",
								Type:       metric.MetricNumber,
								TypeString: "uint64",
								Value:      uint64(50),
								Unit:       "",
							},
							"free": &metric.Metric{
								Name:       "free",
								Dimension:  "none",
								Type:       metric.MetricNumber,
								TypeString: "uint64",
								Value:      uint64(27),
								Unit:       "",
							},
							"used": &metric.Metric{
								Name:       "used",
								Dimension:  "none",
								Type:       metric.MetricNumber,
								TypeString: "uint64",
								Value:      uint64(23),
								Unit:       "",
							},
							"avail": &metric.Metric{
								Name:       "avail",
								Dimension:  "none",
								Type:       metric.MetricNumber,
								TypeString: "uint64",
								Value:      uint64(25),
								Unit:       "",
							},
							"files": &metric.Metric{
								Name:       "files",
								Dimension:  "none",
								Type:       metric.MetricNumber,
								TypeString: "uint64",
								Value:      uint64(26),
								Unit:       "",
							},
							"free_files": &metric.Metric{
								Name:       "free_files",
								Dimension:  "none",
								Type:       metric.MetricNumber,
								TypeString: "uint64",
								Value:      uint64(10),
								Unit:       "",
							},
						},
					},
				},
			},
			expected: &protocol_hostinfo.HostInfoFilesystemResult{
				Metrics: []protocol_hostinfo.HostInfoFilesystemMetrics{
					protocol_hostinfo.HostInfoFilesystemMetrics{
						DirectoryName:  "mount point",
						DeviceName:     "partition device",
						SystemTypeName: "partition fstype",
						Options:        "partition options",
						Available:      25,
						Files:          26,
						Free:           27,
						FreeFiles:      10,
						Total:          50,
						Used:           23,
					},
				},
				Timestamp: utils.NowTimestampMillis(),
			},
		},
		{
			name: "Crs is nil",
			crs:  nil,
			expected: &protocol_hostinfo.HostInfoFilesystemResult{
				Timestamp: utils.NowTimestampMillis(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &hostinfo.HostInfoFilesystem{
				HostInfoBase: protocol_hostinfo.HostInfoBase{},
			}

			got := h.BuildResult(tt.crs)
			assert.Equal(
				t, tt.expected.Metrics,
				got.(*protocol_hostinfo.HostInfoFilesystemResult).Metrics,
				fmt.Sprintf("HostInfoFilesystem.BuildResult() = %v, expected %v", got.(*protocol_hostinfo.HostInfoFilesystemResult).Metrics, tt.expected.Metrics))

		})
	}
}
