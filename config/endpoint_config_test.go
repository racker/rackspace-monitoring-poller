package config_test

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/racker/rackspace-monitoring-poller/config"
)

type endpoint_fields struct {
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
			name:     "Init test",
			expected: &config.EndpointConfig{},
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
	tempList := []string{}
	tests := []struct {
		name        string
		fields      endpoint_fields
		filepath    func() string
		expectedErr bool
	}{
		{
			name:   "Error on file open",
			fields: endpoint_fields{},
			filepath: func() string {
				return "noexiste"
			},
			expectedErr: true,
		},
		{
			name:   "Empty config file",
			fields: endpoint_fields{},
			filepath: func() string {
				f, _ := ioutil.TempFile("", "load_path")
				tempList = append(tempList, f.Name())

				f.Sync()
				defer f.Close()
				return f.Name()
			},
			expectedErr: true,
		},
		{
			name:   "No comments config file",
			fields: endpoint_fields{},
			filepath: func() string {
				f, _ := ioutil.TempFile("", "load_path")
				tempList = append(tempList, f.Name())
				f.Write([]byte("hello\nworld\n"))

				f.Sync()
				defer f.Close()
				return f.Name()
			},
			expectedErr: false,
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
			err := cfg.LoadFromFile(tt.filepath())
			if tt.expectedErr {

			}
			if err != nil {
				if !tt.expectedErr {
					t.Errorf("EndpointConfig.LoadFromFile() error = %v, expectedErr %v", err, tt.expectedErr)
				}
			} else {

			}
		})
	}
}
