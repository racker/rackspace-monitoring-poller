package config_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/racker/rackspace-monitoring-poller/config"
)

func TestNewConfig(t *testing.T) {
	type args struct {
		guid string
	}
	tests := []struct {
		name string
		args args
		want *config.Config
	}{
		{
			name: "HappyPath",
			args: args{
				guid: "some-guid",
			},
			want: &config.Config{
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
			got := config.NewConfig(tt.args.guid)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_LoadFromFile(t *testing.T) {
	type fields struct {
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
	type args struct {
		filepath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
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
			if err := cfg.LoadFromFile(tt.args.filepath); (err != nil) != tt.wantErr {
				t.Errorf("Config.LoadFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_ParseFields(t *testing.T) {
	type fields struct {
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
	type args struct {
		fields []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *config.Config
		err    error
	}{
		{
			name: "Set Monitoring Id",
			fields: fields{
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
			args: args{
				fields: []string{
					"monitoring_id",
					"agentname",
				},
			},
			want: &config.Config{
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
			fields: fields{
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
			args: args{
				fields: []string{
					"monitoring_id",
				},
			},
			want: &config.Config{
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
			fields: fields{
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
			args: args{
				fields: []string{
					"monitoring_token",
					"myawesometoken",
				},
			},
			want: &config.Config{
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
			fields: fields{
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
			args: args{
				fields: []string{
					"monitoring_token",
				},
			},
			want: &config.Config{
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
			fields: fields{
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
			args: args{
				fields: []string{
					"monitoring_endpoints",
					"127.0.0.1,0.0.0.0",
				},
			},
			want: &config.Config{
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
			fields: fields{
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
			args: args{
				fields: []string{
					"monitoring_endpoints",
				},
			},
			want: &config.Config{
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
			fields: fields{
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
			args: args{
				fields: []string{
					"whatiseven",
					"thething",
				},
			},
			want: &config.Config{
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
			err := cfg.ParseFields(tt.args.fields)
			if !reflect.DeepEqual(cfg, tt.want) {
				t.Errorf("TestConfig_ParseFields() = %v, want %v", cfg, tt.want)
			}
			if err != tt.err {
				t.Errorf("TestConfig_ParseFields() error = %v, want %v", err, tt.err)
			}
		})
	}
}

func TestConfig_SetPrivateZones(t *testing.T) {
	type fields struct {
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
	type args struct {
		zones []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *config.Config
	}{
		{
			name: "Set Private zones",
			fields: fields{
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
			args: args{
				zones: []string{
					"zone1",
					"zone2",
				},
			},
			want: &config.Config{
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
			cfg.SetPrivateZones(tt.args.zones)
			if !reflect.DeepEqual(cfg, tt.want) {
				t.Errorf("TestConfig_SetPrivateZones() = %v, want %v", cfg, tt.want)
			}
		})
	}
}

func TestConfig_GetReadDeadline(t *testing.T) {
	type fields struct {
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
	type args struct {
		offset time.Duration
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   time.Time
	}{
	// TODO: Add test cases.
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
			if got := cfg.GetReadDeadline(tt.args.offset); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Config.GetReadDeadline() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetWriteDeadline(t *testing.T) {
	type fields struct {
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
	type args struct {
		offset time.Duration
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   time.Time
	}{
	// TODO: Add test cases.
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
			if got := cfg.GetWriteDeadline(tt.args.offset); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Config.GetWriteDeadline() = %v, want %v", got, tt.want)
			}
		})
	}
}
