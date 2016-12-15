package config_test

import (
	"errors"
	"io/ioutil"
	"os"
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
		filepath    string
		osopen      func(fp string) (*os.File, error)
		expectedErr bool
	}{
		{
			name:     "Error on file open",
			fields:   endpoint_fields{},
			filepath: "testpath",
			osopen: func(fp string) (*os.File, error) {
				return nil, errors.New("Why?!?")
			},
			expectedErr: true,
		},
		{
			name:     "Bad config file",
			fields:   endpoint_fields{},
			filepath: "testpath",
			osopen: func(fp string) (*os.File, error) {
				f, err := ioutil.TempFile("", "load_path")
				tempList = append(tempList, f.Name())
				f.Write([]byte("hello\nworld\n"))

				f.Sync()
				defer f.Close()
				return f, err
			},
			expectedErr: true,
		},
		{
			name:     "No comments config file",
			fields:   endpoint_fields{},
			filepath: "testpath",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock os open
			osopen := config.OsOpen
			config.OsOpen = tt.osopen
			defer func() { config.OsOpen = osopen }()

			cfg := &config.EndpointConfig{
				CertFile:        tt.fields.CertFile,
				KeyFile:         tt.fields.KeyFile,
				BindAddr:        tt.fields.BindAddr,
				StatsDAddr:      tt.fields.StatsDAddr,
				AgentsConfigDir: tt.fields.AgentsConfigDir,
			}
			err := cfg.LoadFromFile(tt.filepath)
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
