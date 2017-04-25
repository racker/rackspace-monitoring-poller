package config_test

import (
	"testing"
	"time"

	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/stretchr/testify/assert"
)

type configFields struct {
	UseSrv            bool
	SrvQueries        []string
	Addresses         []string
	AgentId           string
	AgentName         string
	Features          []config.Feature
	Guid              string
	BundleVersion     string
	ProcessVersion    string
	Token             string
	ZoneIds           []string
	TimeoutRead       time.Duration
	TimeoutWrite      time.Duration
	TimeoutPrepareEnd time.Duration
}

func getConfigFields() configFields {
	return configFields{
		UseSrv:            true,
		AgentName:         "remote_poller",
		AgentId:           "-poller-",
		ProcessVersion:    "dev",
		BundleVersion:     "dev",
		Guid:              "some-guid",
		TimeoutRead:       time.Duration(10 * time.Second),
		TimeoutWrite:      time.Duration(10 * time.Second),
		TimeoutPrepareEnd: config.DefaultTimeoutPrepareEnd,
		Token:             "",
		Features:          make([]config.Feature, 0),
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
				AgentName:              "remote_poller",
				AgentId:                "-poller-",
				ProcessVersion:         "dev",
				BundleVersion:          "dev",
				Guid:                   "some-guid",
				ReconnectMinBackoff:    config.DefaultReconnectMinBackoff,
				ReconnectMaxBackoff:    config.DefaultReconnectMaxBackoff,
				ReconnectFactorBackoff: config.DefaultReconnectFactorBackoff,
				TimeoutRead:            time.Duration(10 * time.Second),
				TimeoutWrite:           time.Duration(10 * time.Second),
				TimeoutPrepareEnd:      config.DefaultTimeoutPrepareEnd,
				TimeoutAuth:            config.DefaultTimeoutAuth,
				Token:                  "",
				Features:               make([]config.Feature, 0),
			},
		},
		{
			name:       "UseStaging",
			guid:       "some-guid-via-staging",
			useStaging: true,
			expected: &config.Config{
				UseSrv:     true,
				UseStaging: true,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.stage.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.stage.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.stage.monitoring.api.rackspacecloud.com",
				},
				AgentName:              "remote_poller",
				AgentId:                "-poller-",
				ProcessVersion:         "dev",
				BundleVersion:          "dev",
				Guid:                   "some-guid-via-staging",
				ReconnectMinBackoff:    config.DefaultReconnectMinBackoff,
				ReconnectMaxBackoff:    config.DefaultReconnectMaxBackoff,
				ReconnectFactorBackoff: config.DefaultReconnectFactorBackoff,
				TimeoutRead:            time.Duration(10 * time.Second),
				TimeoutWrite:           time.Duration(10 * time.Second),
				TimeoutPrepareEnd:      config.DefaultTimeoutPrepareEnd,
				TimeoutAuth:            config.DefaultTimeoutAuth,
				Token:                  "",
				Features:               make([]config.Feature, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.NewConfig(tt.guid, tt.useStaging, nil)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestConfig_LoadFromFile(t *testing.T) {
	tests := []struct {
		name        string
		filepath    string
		expectedErr bool
		expected    *config.Config
	}{
		{
			name:        "Error on file open",
			filepath:    "noexiste",
			expectedErr: true,
		},
		{
			name:        "No comments config file",
			filepath:    "testdata/no-comments-config-file.txt",
			expected:    config.NewConfig("1-2-3", false, nil),
			expectedErr: false,
		},
		{
			name:        "With comments in config file",
			filepath:    "testdata/with-comments-config-file.txt",
			expected:    config.NewConfig("1-2-3", false, nil),
			expectedErr: false,
		},
		{
			name: "With snet region",
			expected: expectedConfigWithSrvQueries("1-2-3", false, "dfw", []string{
				"_monitoringagent._tcp.snet-dfw-region0.prod.monitoring.api.rackspacecloud.com",
				"_monitoringagent._tcp.snet-dfw-region1.prod.monitoring.api.rackspacecloud.com",
				"_monitoringagent._tcp.snet-dfw-region2.prod.monitoring.api.rackspacecloud.com",
			}),
			filepath:    "testdata/with-snet-config-file.txt",
			expectedErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cfg := config.NewConfig("1-2-3", false, nil)
			err := cfg.LoadFromFile(tt.filepath)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, cfg)
			}

		})
	}
}

func expectedConfigWithSrvQueries(guid string, useStaging bool, snetRegion string, srvQueries []string) *config.Config {
	cfg := config.NewConfig(guid, useStaging, nil)
	cfg.SnetRegion = snetRegion
	cfg.SrvQueries = srvQueries
	return cfg
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
				UseSrv:         true,
				AgentName:      "remote_poller",
				AgentId:        "agentname",
				ProcessVersion: "dev",
				BundleVersion:  "dev",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]config.Feature, 0),
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
				UseSrv:         true,
				AgentName:      "remote_poller",
				AgentId:        "-poller-",
				ProcessVersion: "dev",
				BundleVersion:  "dev",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]config.Feature, 0),
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
				UseSrv:         true,
				AgentName:      "remote_poller",
				AgentId:        "-poller-",
				ProcessVersion: "dev",
				BundleVersion:  "dev",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "myawesometoken",
				Features:       make([]config.Feature, 0),
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
				UseSrv:         true,
				AgentName:      "remote_poller",
				AgentId:        "-poller-",
				ProcessVersion: "dev",
				BundleVersion:  "dev",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]config.Feature, 0),
			},
			expectedErr: true,
		},
		{
			name:   "Set Monitoring Endpoint",
			fields: getConfigFields(),
			args: []string{
				"monitoring_endpoints",
				"127.0.0.1,0.0.0.0",
			},
			expected: &config.Config{
				UseSrv:         false,
				AgentName:      "remote_poller",
				AgentId:        "-poller-",
				ProcessVersion: "dev",
				BundleVersion:  "dev",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Addresses: []string{
					"127.0.0.1",
					"0.0.0.0",
				},
				Features: make([]config.Feature, 0),
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
				UseSrv:         true,
				AgentName:      "remote_poller",
				AgentId:        "-poller-",
				ProcessVersion: "dev",
				BundleVersion:  "dev",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]config.Feature, 0),
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
				UseSrv:         true,
				AgentName:      "remote_poller",
				AgentId:        "-poller-",
				ProcessVersion: "dev",
				BundleVersion:  "dev",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]config.Feature, 0),
			},
			expectedErr: false,
		},
		{
			name:   "ValidSnet",
			fields: getConfigFields(),
			args: []string{
				"monitoring_snet_region", "iad",
			},
			expected: &config.Config{
				UseSrv:         true,
				AgentName:      "remote_poller",
				AgentId:        "-poller-",
				ProcessVersion: "dev",
				BundleVersion:  "dev",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]config.Feature, 0),
				SnetRegion:     "iad",
			},
			expectedErr: false,
		},
		{
			name:   "InvalidSnet",
			fields: getConfigFields(),
			args: []string{
				"monitoring_snet_region", "neverneverland",
			},
			expected: &config.Config{
				UseSrv:         true,
				AgentName:      "remote_poller",
				AgentId:        "-poller-",
				ProcessVersion: "dev",
				BundleVersion:  "dev",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]config.Feature, 0),
				SnetRegion:     "",
			},
			expectedErr: true,
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
			err := cfg.ParseFields(cfg.DefineConfigEntries(), tt.args)
			assert.Equal(t, tt.expected, cfg)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
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
				UseSrv:         true,
				AgentName:      "remote_poller",
				AgentId:        "-poller-",
				ProcessVersion: "dev",
				BundleVersion:  "dev",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]config.Feature, 0),
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
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

func TestConfig_PostProcess(t *testing.T) {
	tests := []struct {
		Name               string
		Given              config.Config
		ExpectedSrvQueries []string
	}{
		{
			Name: "SnetProduction",
			Given: config.Config{
				UseStaging: false,
				SnetRegion: "iad",
			},
			ExpectedSrvQueries: []string{
				"_monitoringagent._tcp.snet-iad-region0.prod.monitoring.api.rackspacecloud.com",
				"_monitoringagent._tcp.snet-iad-region1.prod.monitoring.api.rackspacecloud.com",
				"_monitoringagent._tcp.snet-iad-region2.prod.monitoring.api.rackspacecloud.com",
			},
		},
		{
			Name: "SnetStaging",
			Given: config.Config{
				UseStaging: true,
				SnetRegion: "ord",
			},
			ExpectedSrvQueries: []string{
				"_monitoringagent._tcp.snet-ord-region0.stage.monitoring.api.rackspacecloud.com",
				"_monitoringagent._tcp.snet-ord-region1.stage.monitoring.api.rackspacecloud.com",
				"_monitoringagent._tcp.snet-ord-region2.stage.monitoring.api.rackspacecloud.com",
			},
		},
		{
			Name: "NoSnet",
			Given: config.Config{
				SnetRegion: "",
			},
			ExpectedSrvQueries: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			var processed config.Config = tt.Given
			processed.PostProcess()
			assert.Equal(t, tt.ExpectedSrvQueries, processed.SrvQueries)
		})
	}
}
