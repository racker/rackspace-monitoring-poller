package poller_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"net/url"

	"time"

	"crypto/x509"

	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/poller"
	"github.com/stretchr/testify/assert"
)

func TestNewConnectionStream(t *testing.T) {
	testConfig := &config.Config{
		AgentId: "awesome agent",
	}
	multipleZoneIdsConfig := &config.Config{
		AgentId: "awesome agent",
		ZoneIds: []string{"zone one", "zone two"},
	}
	tests := []struct {
		name     string
		config   *config.Config
		rootCA   *x509.CertPool
		expected *config.Config
	}{
		{
			name:     "Happy path",
			config:   testConfig,
			rootCA:   x509.NewCertPool(),
			expected: testConfig,
		},
		{
			name:     "Multiple ZoneIds",
			config:   multipleZoneIdsConfig,
			rootCA:   x509.NewCertPool(),
			expected: multipleZoneIdsConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := poller.NewConnectionStream(tt.config, tt.rootCA)
			//assert that configs are the same
			assert.Equal(t, tt.expected, got.GetConfig())
		})
	}
}

func TestConnectionStream_Register(t *testing.T) {
	tests := []struct {
		name        string
		queryString string
		conn        poller.Connection
		cs          poller.ConnectionStream
		expectedErr bool
	}{
		{
			name:        "Happy path",
			queryString: "test-query",
			conn:        &poller.EleConnection{},
			cs:          poller.NewConnectionStream(&config.Config{}, nil),
			expectedErr: false,
		},
		{
			name:        "Empty stream",
			queryString: "test-query",
			conn:        &poller.EleConnection{},
			cs:          &poller.EleConnectionStream{},
			expectedErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotContains(t, tt.cs.GetConnections(), tt.queryString)
			if tt.expectedErr {
				assert.Error(t, tt.cs.RegisterConnection(tt.queryString, tt.conn))
				assert.NotContains(t, tt.cs.GetConnections(), tt.queryString)

			} else {
				assert.NoError(t, tt.cs.RegisterConnection(tt.queryString, tt.conn))
				assert.Contains(t, tt.cs.GetConnections(), tt.queryString)

			}
		})
	}
}

func TestConnectionStream_Connect(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(staticResponse))
	defer ts.Close()

	tests := []struct {
		name          string
		addresses     func() []string
		serverQueries func() []string
		useSrv        bool
	}{
		{
			name: "Happy path",
			addresses: func() []string {
				testURL, _ := url.Parse(ts.URL)
				return []string{testURL.Host}
			},
			serverQueries: func() []string {
				return []string{}
			},
			useSrv: false,
		},
		{
			name: "Use service",
			addresses: func() []string {
				return []string{}
			},
			serverQueries: func() []string {
				return []string{"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com"}
			},
			useSrv: true,
		},
		{
			name: "Invalid service query",
			addresses: func() []string {
				return []string{}
			},
			serverQueries: func() []string {
				return []string{"magic"}
			},
			useSrv: true,
		},
		{
			name: "Invalid url",
			addresses: func() []string {
				return []string{"invalid-url:1234"}
			},
			serverQueries: func() []string {
				return []string{}
			},
			useSrv: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := poller.NewConnectionStream(&config.Config{
				UseSrv:     tt.useSrv,
				Addresses:  tt.addresses(),
				SrvQueries: tt.serverQueries(),
			}, nil)
			go cs.Connect()

			go func() {
				time.Sleep(25 * time.Millisecond)
				cs.Stop()
			}()

			select {
			case <-cs.StopNotify():
				assert.True(t, true, "blah")
			case <-time.After(1 * time.Minute):
				assert.Fail(t, "fail")
			}

		})
	}
}
