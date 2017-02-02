package config_test

import (
	"reflect"
	"testing"

	"github.com/racker/rackspace-monitoring-poller/config"
)

type endpointFields struct {
	CertFile        string
	KeyFile         string
	BindAddr        string
	StatsDAddr      string
	AgentsConfigDir string
}

func TestNewEndpointConfig(t *testing.T) {
	tests := []struct {
		name     string
		expected *config.EndpointConfig
	}{
		{
			name: "Init test",
			expected: &config.EndpointConfig{
				PrepareBlockSize: 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := config.NewEndpointConfig(); !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NewEndpointConfig() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestEndpointConfig_LoadFromFile(t *testing.T) {
	tests := []struct {
		name        string
		fields      endpointFields
		filepath    string
		expectedErr bool
		expected    *config.EndpointConfig
	}{
		{
			name:        "Error on file open",
			fields:      endpointFields{},
			filepath:    "noexiste",
			expectedErr: true,
			expected:    &config.EndpointConfig{},
		},
		{
			name:        "Empty config file",
			fields:      endpointFields{},
			filepath:    "testdata/empty-file.txt",
			expectedErr: true,
			expected:    &config.EndpointConfig{},
		},
		{
			name:        "Invalid json config file",
			fields:      endpointFields{},
			filepath:    "testdata/no-comments-config-file.txt",
			expectedErr: true,
			expected:    &config.EndpointConfig{},
		},
		{
			name:        "Valid json config file",
			fields:      endpointFields{},
			filepath:    "testdata/valid-endpoint-config.json",
			expectedErr: false,
			expected: &config.EndpointConfig{
				CertFile: "mycertfile",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.EndpointConfig{
				CertFile:        tt.fields.CertFile,
				KeyFile:         tt.fields.KeyFile,
				BindAddr:        tt.fields.BindAddr,
				StatsDAddr:      tt.fields.StatsDAddr,
				AgentsConfigDir: tt.fields.AgentsConfigDir,
			}
			err := cfg.LoadFromFile(tt.filepath)
			if tt.expectedErr {
				if err == nil {
					t.Error("EndpointConfig.LoadFromFile() expected error")
				}
			} else {
				if err != nil {
					t.Errorf("EndpointConfig.LoadFromFile() error = %v, expectedErr %v", err, tt.expectedErr)
				}
			}
			//check cfg
			if !reflect.DeepEqual(cfg, tt.expected) {
				t.Errorf("EndpointConfig.LoadFromFile() = %v, expected %v", cfg, tt.expected)
			}
		})
	}
}
