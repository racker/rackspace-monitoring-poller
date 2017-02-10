package config_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/stretchr/testify/assert"
)

type configFields struct {
	UseSrv         bool
	SrvQueries     []string
	Addresses      []string
	AgentId        string
	AgentName      string
	Features       []map[string]string
	Guid           string
	BundleVersion  string
	ProcessVersion string
	Token          string
	ZoneIds        []string
	TimeoutRead    time.Duration
	TimeoutWrite   time.Duration
}

func getConfigFields() configFields {
	return configFields{
		UseSrv: true,
		SrvQueries: []string{
			"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
			"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
			"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
		},
		AgentName:      "remote_poller",
		ProcessVersion: "dev",
		BundleVersion:  "dev",
		Guid:           "some-guid",
		TimeoutRead:    time.Duration(10 * time.Second),
		TimeoutWrite:   time.Duration(10 * time.Second),
		Token:          "",
		Features:       make([]map[string]string, 0),
	}
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name       string
		guid       string
		useStaging bool
		expected   *config.Config
	}{
		{
			name:       "HappyPath",
			guid:       "some-guid",
			useStaging: false,
			expected: &config.Config{
				UseSrv: true,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
				},
				AgentName:      "remote_poller",
				ProcessVersion: "dev",
				BundleVersion:  "dev",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
		},
		{
			name:       "UseStating",
			guid:       "some-guid-via-staging",
			useStaging: true,
			expected: &config.Config{
				UseSrv: true,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.stage.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.stage.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.stage.monitoring.api.rackspacecloud.com",
				},
				AgentName:      "remote_poller",
				ProcessVersion: "dev",
				BundleVersion:  "dev",
				Guid:           "some-guid-via-staging",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.NewConfig(tt.guid, tt.useStaging)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestConfig_LoadFromFile(t *testing.T) {
	tests := []struct {
		name        string
		fields      configFields
		filepath    string
		expectedErr bool
	}{
		{
			name:        "Error on file open",
			fields:      configFields{},
			filepath:    "noexiste",
			expectedErr: true,
		},
		{
			name:        "No comments config file",
			fields:      configFields{},
			filepath:    "testdata/no-comments-config-file.txt",
			expectedErr: false,
		},
		{
			name:        "With comments in config file",
			fields:      configFields{},
			filepath:    "testdata/with-comments-config-file.txt",
			expectedErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cfg := &config.Config{
				UseSrv:         tt.fields.UseSrv,
				SrvQueries:     tt.fields.SrvQueries,
				Addresses:      tt.fields.Addresses,
				AgentId:        tt.fields.AgentId,
				AgentName:      tt.fields.AgentName,
				Features:       tt.fields.Features,
				Guid:           tt.fields.Guid,
				BundleVersion:  tt.fields.BundleVersion,
				ProcessVersion: tt.fields.ProcessVersion,
				Token:          tt.fields.Token,
				ZoneIds:        tt.fields.ZoneIds,
				TimeoutRead:    tt.fields.TimeoutRead,
				TimeoutWrite:   tt.fields.TimeoutWrite,
			}
			if err := cfg.LoadFromFile(tt.filepath); (err != nil) != tt.expectedErr {
				t.Errorf("Config.LoadFromFile() error = %v, expectedErr %v", err, tt.expectedErr)
			}
		})
	}
}

func TestConfig_ParseFields(t *testing.T) {
	tests := []struct {
		name        string
		fields      configFields
		args        []string
		expected    *config.Config
		expectedErr bool
	}{
		{
			name:   "Set Monitoring Id",
			fields: getConfigFields(),
			args: []string{
				"monitoring_id",
				"agentname",
			},
			expected: &config.Config{
				UseSrv: true,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
				},
				AgentName:      "remote_poller",
				AgentId:        "agentname",
				ProcessVersion: "dev",
				BundleVersion:  "dev",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
			expectedErr: false,
		},
		{
			name:   "Set Monitoring Id without agent id",
			fields: getConfigFields(),
			args: []string{
				"monitoring_id",
			},
			expected: &config.Config{
				UseSrv: true,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
				},
				AgentName:      "remote_poller",
				ProcessVersion: "dev",
				BundleVersion:  "dev",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
			expectedErr: true,
		},
		{
			name:   "Set Monitoring Token",
			fields: getConfigFields(),
			args: []string{
				"monitoring_token",
				"myawesometoken",
			},
			expected: &config.Config{
				UseSrv: true,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
				},
				AgentName:      "remote_poller",
				ProcessVersion: "dev",
				BundleVersion:  "dev",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "myawesometoken",
				Features:       make([]map[string]string, 0),
			},
			expectedErr: false,
		},
		{
			name:   "Set Monitoring Token without token",
			fields: getConfigFields(),
			args: []string{
				"monitoring_token",
			},
			expected: &config.Config{
				UseSrv: true,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
				},
				AgentName:      "remote_poller",
				ProcessVersion: "dev",
				BundleVersion:  "dev",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
			expectedErr: true,
		},
		{
			name:   "Set Monitoring Endpoint",
			fields: getConfigFields(),
			args: []string{
				"monitoring_endpoints",
				"127.dev,0.0.0.0",
			},
			expected: &config.Config{
				UseSrv: false,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
				},
				AgentName:      "remote_poller",
				ProcessVersion: "dev",
				BundleVersion:  "dev",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Addresses: []string{
					"127.dev",
					"0.0.0.0",
				},
				Features: make([]map[string]string, 0),
			},
			expectedErr: false,
		},
		{
			name:   "Set Monitoring Endpoint without addresses",
			fields: getConfigFields(),
			args: []string{
				"monitoring_endpoints",
			},
			expected: &config.Config{
				UseSrv: true,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
				},
				AgentName:      "remote_poller",
				ProcessVersion: "dev",
				BundleVersion:  "dev",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
			expectedErr: true,
		},
		{
			name:   "Randomness",
			fields: getConfigFields(),
			args: []string{
				"whatiseven",
				"thething",
			},
			expected: &config.Config{
				UseSrv: true,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
				},
				AgentName:      "remote_poller",
				ProcessVersion: "dev",
				BundleVersion:  "dev",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
			expectedErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				UseSrv:         tt.fields.UseSrv,
				SrvQueries:     tt.fields.SrvQueries,
				Addresses:      tt.fields.Addresses,
				AgentId:        tt.fields.AgentId,
				AgentName:      tt.fields.AgentName,
				Features:       tt.fields.Features,
				Guid:           tt.fields.Guid,
				BundleVersion:  tt.fields.BundleVersion,
				ProcessVersion: tt.fields.ProcessVersion,
				Token:          tt.fields.Token,
				ZoneIds:        tt.fields.ZoneIds,
				TimeoutRead:    tt.fields.TimeoutRead,
				TimeoutWrite:   tt.fields.TimeoutWrite,
			}
			err := cfg.ParseFields(tt.args)
			if !reflect.DeepEqual(cfg, tt.expected) {
				t.Errorf("TestConfig_ParseFields() = %v, expected %v", cfg, tt.expected)
			}
			if tt.expectedErr {
				if err == nil {
					t.Error("TestConfig_ParseFields() expected error")
				}
			} else {
				if err != nil {
					t.Errorf("TestConfig_ParseFields() error = %v, expectedErr %v", err, tt.expectedErr)
				}
			}
		})
	}
}

func TestConfig_SetPrivateZones(t *testing.T) {

	tests := []struct {
		name     string
		fields   configFields
		zones    []string
		expected *config.Config
	}{
		{
			name:   "Set Private zones",
			fields: getConfigFields(),
			zones: []string{
				"zone1",
				"zone2",
			},
			expected: &config.Config{
				UseSrv: true,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
				},
				AgentName:      "remote_poller",
				ProcessVersion: "dev",
				BundleVersion:  "dev",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
				ZoneIds: []string{
					"zone1",
					"zone2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				UseSrv:         tt.fields.UseSrv,
				SrvQueries:     tt.fields.SrvQueries,
				Addresses:      tt.fields.Addresses,
				AgentId:        tt.fields.AgentId,
				AgentName:      tt.fields.AgentName,
				Features:       tt.fields.Features,
				Guid:           tt.fields.Guid,
				BundleVersion:  tt.fields.BundleVersion,
				ProcessVersion: tt.fields.ProcessVersion,
				Token:          tt.fields.Token,
				ZoneIds:        tt.fields.ZoneIds,
				TimeoutRead:    tt.fields.TimeoutRead,
				TimeoutWrite:   tt.fields.TimeoutWrite,
			}
			cfg.SetPrivateZones(tt.zones)
			if !reflect.DeepEqual(cfg, tt.expected) {
				t.Errorf("TestConfig_SetPrivateZones() = %v, expected %v", cfg, tt.expected)
			}
		})
	}
}
