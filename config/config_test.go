package config_test

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"errors"

	"github.com/racker/rackspace-monitoring-poller/config"
)

type myFileInfo os.FileInfo

type config_fields struct {
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
	PrivateZones   []string
	TimeoutRead    time.Duration
	TimeoutWrite   time.Duration
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name     string
		guid     string
		expected *config.Config
	}{
		{
			name: "HappyPath",
			guid: "some-guid",
			expected: &config.Config{
				UseSrv: true,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
				},
				AgentName:      "remote_poller",
				ProcessVersion: "0.0.1",
				BundleVersion:  "0.0.1",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.NewConfig(tt.guid)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NewConfig() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestConfig_LoadFromFile(t *testing.T) {
	tempList := []string{}
	tests := []struct {
		name        string
		fields      config_fields
		filepath    string
		osstat      func(fp string) (os.FileInfo, error)
		osopen      func(fp string) (*os.File, error)
		expectedErr bool
	}{
		{
			name:     "Error on fileinfo",
			fields:   config_fields{},
			filepath: "testpath",
			osstat: func(fp string) (os.FileInfo, error) {
				return nil, errors.New("We fail everything")
			},
			osopen:      func(fp string) (*os.File, error) { return nil, nil },
			expectedErr: true,
		},
		{
			name:     "Error on file open",
			fields:   config_fields{},
			filepath: "testpath",
			osstat: func(fp string) (os.FileInfo, error) {
				return nil, nil
			},
			osopen: func(fp string) (*os.File, error) {
				return nil, errors.New("Why?!?")
			},
			expectedErr: true,
		},
		{
			name:     "No comments config file",
			fields:   config_fields{},
			filepath: "testpath",
			osstat:   func(fp string) (os.FileInfo, error) { return nil, nil },
			osopen: func(fp string) (*os.File, error) {
				f, _ := ioutil.TempFile("", "load_path")
				tempList = append(tempList, f.Name())
				f.Write([]byte("hello\nworld\n"))

				f.Sync()
				defer f.Close()
				return os.Open(f.Name())
			},
			expectedErr: false,
		},
		{
			name:     "With comments in config file",
			fields:   config_fields{},
			filepath: "testpath",
			osstat:   func(fp string) (os.FileInfo, error) { return nil, nil },
			osopen: func(fp string) (*os.File, error) {
				f, _ := ioutil.TempFile("", "load_path")
				tempList = append(tempList, f.Name())
				f.Write([]byte("hello\n#world\nrackspace\n"))

				f.Sync()
				defer f.Close()
				return os.Open(f.Name())
			},
			expectedErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock os stat
			osstat := config.OsStat
			config.OsStat = tt.osstat
			defer func() { config.OsStat = osstat }()
			// mock os open
			osopen := config.OsOpen
			config.OsOpen = tt.osopen
			defer func() { config.OsOpen = osopen }()

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
				PrivateZones:   tt.fields.PrivateZones,
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
		name     string
		fields   config_fields
		args     []string
		expected *config.Config
		err      error
	}{
		{
			name: "Set Monitoring Id",
			fields: config_fields{
				UseSrv: true,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
				},
				AgentName:      "remote_poller",
				ProcessVersion: "0.0.1",
				BundleVersion:  "0.0.1",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
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
				ProcessVersion: "0.0.1",
				BundleVersion:  "0.0.1",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
			err: nil,
		},
		{
			name: "Set Monitoring Id without agent id",
			fields: config_fields{
				UseSrv: true,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
				},
				AgentName:      "remote_poller",
				ProcessVersion: "0.0.1",
				BundleVersion:  "0.0.1",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
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
				ProcessVersion: "0.0.1",
				BundleVersion:  "0.0.1",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
			err: config.BadConfig{Details: "Invalid fields length"},
		},
		{
			name: "Set Monitoring Token",
			fields: config_fields{
				UseSrv: true,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
				},
				AgentName:      "remote_poller",
				ProcessVersion: "0.0.1",
				BundleVersion:  "0.0.1",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
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
				ProcessVersion: "0.0.1",
				BundleVersion:  "0.0.1",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "myawesometoken",
				Features:       make([]map[string]string, 0),
			},
			err: nil,
		},
		{
			name: "Set Monitoring Token without token",
			fields: config_fields{
				UseSrv: true,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
				},
				AgentName:      "remote_poller",
				ProcessVersion: "0.0.1",
				BundleVersion:  "0.0.1",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
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
				ProcessVersion: "0.0.1",
				BundleVersion:  "0.0.1",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
			err: config.BadConfig{Details: "Invalid fields length"},
		},
		{
			name: "Set Monitoring Endpoint",
			fields: config_fields{
				UseSrv: true,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
				},
				AgentName:      "remote_poller",
				ProcessVersion: "0.0.1",
				BundleVersion:  "0.0.1",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
			args: []string{
				"monitoring_endpoints",
				"127.0.0.1,0.0.0.0",
			},
			expected: &config.Config{
				UseSrv: false,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
				},
				AgentName:      "remote_poller",
				ProcessVersion: "0.0.1",
				BundleVersion:  "0.0.1",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Addresses: []string{
					"127.0.0.1",
					"0.0.0.0",
				},
				Features: make([]map[string]string, 0),
			},
			err: nil,
		},
		{
			name: "Set Monitoring Endpoint without addresses",
			fields: config_fields{
				UseSrv: true,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
				},
				AgentName:      "remote_poller",
				ProcessVersion: "0.0.1",
				BundleVersion:  "0.0.1",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
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
				ProcessVersion: "0.0.1",
				BundleVersion:  "0.0.1",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
			err: config.BadConfig{Details: "Invalid fields length"},
		},
		{
			name: "Randomness",
			fields: config_fields{
				UseSrv: true,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
				},
				AgentName:      "remote_poller",
				ProcessVersion: "0.0.1",
				BundleVersion:  "0.0.1",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
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
				ProcessVersion: "0.0.1",
				BundleVersion:  "0.0.1",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
			err: nil,
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
				PrivateZones:   tt.fields.PrivateZones,
				TimeoutRead:    tt.fields.TimeoutRead,
				TimeoutWrite:   tt.fields.TimeoutWrite,
			}
			err := cfg.ParseFields(tt.args)
			if !reflect.DeepEqual(cfg, tt.expected) {
				t.Errorf("TestConfig_ParseFields() = %v, expected %v", cfg, tt.expected)
			}
			if err != tt.err {
				t.Errorf("TestConfig_ParseFields() error = %v, expected %v", err, tt.err)
			}
		})
	}
}

func TestConfig_SetPrivateZones(t *testing.T) {

	tests := []struct {
		name     string
		fields   config_fields
		zones    []string
		expected *config.Config
	}{
		{
			name: "Set Private zones",
			fields: config_fields{
				UseSrv: true,
				SrvQueries: []string{
					"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.ord1.prod.monitoring.api.rackspacecloud.com",
					"_monitoringagent._tcp.lon3.prod.monitoring.api.rackspacecloud.com",
				},
				AgentName:      "remote_poller",
				ProcessVersion: "0.0.1",
				BundleVersion:  "0.0.1",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
			},
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
				ProcessVersion: "0.0.1",
				BundleVersion:  "0.0.1",
				Guid:           "some-guid",
				TimeoutRead:    time.Duration(10 * time.Second),
				TimeoutWrite:   time.Duration(10 * time.Second),
				Token:          "",
				Features:       make([]map[string]string, 0),
				PrivateZones: []string{
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
				PrivateZones:   tt.fields.PrivateZones,
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
