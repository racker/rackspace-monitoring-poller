package hostinfo_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/racker/rackspace-monitoring-poller/hostinfo"
	protocol "github.com/racker/rackspace-monitoring-poller/protocol/hostinfo"
)

func TestNewHostInfo(t *testing.T) {
	tests := []struct {
		name      string
		rawParams json.RawMessage
		expected  hostinfo.HostInfo
	}{
		{
			name:      "Memory info",
			rawParams: []byte(`{"type": "MEMORY"}`),
			expected: &hostinfo.HostInfoMemory{
				HostInfoBase: protocol.HostInfoBase{
					Type: protocol.Memory,
				},
			},
		},
		{
			name:      "Cpu info",
			rawParams: []byte(`{"type": "CPU"}`),
			expected: &hostinfo.HostInfoCpu{
				HostInfoBase: protocol.HostInfoBase{
					Type: protocol.Cpu,
				},
			},
		},
		{
			name:      "Filesystem info",
			rawParams: []byte(`{"type": "FILESYSTEM"}`),
			expected: &hostinfo.HostInfoFilesystem{
				HostInfoBase: protocol.HostInfoBase{
					Type: protocol.Filesystem,
				},
			},
		},
		{
			name:      "System info",
			rawParams: []byte(`{"type": "SYSTEM"}`),
			expected: &hostinfo.HostInfoSystem{
				HostInfoBase: protocol.HostInfoBase{
					Type: protocol.System,
				},
			},
		},
		{
			name:      "Processes info",
			rawParams: []byte(`{"type": "PROCS"}`),
			expected: &hostinfo.HostInfoProcesses{
				HostInfoBase: protocol.HostInfoBase{
					Type: protocol.Processes,
				},
			},
		},
		{
			name:      "Fake info",
			rawParams: []byte(`{"type": "FAKE"}`),
			expected:  nil,
		},
		{
			name:      "Error",
			rawParams: []byte(`{"aint","even","json"}`),
			expected:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hostinfo.NewHostInfo(tt.rawParams); !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NewHostInfo() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
