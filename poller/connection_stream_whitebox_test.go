package poller

import (
	"sync"
	"testing"

	"context"

	"log"

	"time"

	"math"

	"github.com/golang/mock/gomock"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/stretchr/testify/assert"
)

func TestConnectionStream_GetConfig(t *testing.T) {
	testConfig := &config.Config{
		AgentId: "test-agent-id",
	}
	tests := []struct {
		name     string
		cs       *ConnectionStream
		expected *config.Config
	}{
		{
			name: "Happy path",
			cs: &ConnectionStream{
				config: testConfig,
			},
			expected: testConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cs.GetConfig()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestConnectionStream_Stop(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConn := NewMockConnectionInterface(mockCtrl)

	tests := []struct {
		name        string
		ctx         context.Context
		stopCh      chan struct{}
		config      *config.Config
		connsMu     sync.Mutex
		conns       map[string]ConnectionInterface
		wg          sync.WaitGroup
		scheduler   *Scheduler
		expectedErr bool
	}{
		{
			name:   "Happy path - one connection",
			ctx:    context.Background(),
			config: &config.Config{},
			conns: map[string]ConnectionInterface{
				"test-query": mockConn,
			},
			expectedErr: false,
		},
		{
			name:   "Happy path - two connections",
			ctx:    context.Background(),
			config: &config.Config{},
			conns: map[string]ConnectionInterface{
				"test-query":         mockConn,
				"another-test-query": mockConn,
			},
			expectedErr: false,
		},
		{
			name:        "No connections",
			ctx:         context.Background(),
			config:      &config.Config{},
			conns:       map[string]ConnectionInterface{},
			expectedErr: false,
		},
		{
			name:        "Connections set to nil",
			ctx:         context.Background(),
			config:      &config.Config{},
			conns:       nil,
			expectedErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &ConnectionStream{
				ctx:       tt.ctx,
				stopCh:    tt.stopCh,
				config:    tt.config,
				connsMu:   tt.connsMu,
				conns:     tt.conns,
				wg:        tt.wg,
				scheduler: tt.scheduler,
			}
			if tt.expectedErr {
				log.Println("aqui")
				mockConn.EXPECT().Close().Times(0)
			} else {
				log.Println(len(tt.conns))
				mockConn.EXPECT().Close().Times(len(tt.conns))
			}
			go cs.Stop()
			// wait for 25 milliseconds for the test to run through everything
			// WARNING: this could be flaky; however, it passed 100 iterations locally
			// this is something to watch
			time.Sleep(25 * time.Millisecond)
		})
	}
}

func TestConnection_GetScheduler(t *testing.T) {
	cs := NewConnectionStream(config.NewConfig("test-guid", false), nil)
	assert.Equal(t, cs.(*ConnectionStream).scheduler, cs.GetScheduler())
}

func TestConnectionStream_SendMetrics(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockSession := NewMockSessionInterface(mockCtrl)

	tests := []struct {
		name        string
		ctx         context.Context
		stopCh      chan struct{}
		config      *config.Config
		connsMu     sync.Mutex
		conns       map[string]ConnectionInterface
		wg          sync.WaitGroup
		scheduler   *Scheduler
		crs         *check.CheckResultSet
		expectedErr bool
	}{
		{
			name:   "Happy path - one session",
			ctx:    context.Background(),
			config: &config.Config{},
			conns: map[string]ConnectionInterface{
				"test-query": &Connection{
					session: mockSession,
				},
			},
			crs: &check.CheckResultSet{
				Check: check.NewCheck([]byte(`{
	  "id":"chPzAHTTP",
	  "zone_id":"pzA",
	  "details":{"url":"localhost"},
	  "type":"remote.http",
	  "timeout":15,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"",
	  "disabled":false
				}`)),
			},
			expectedErr: false,
		},
		{
			name:   "Happy path - two connections",
			ctx:    context.Background(),
			config: &config.Config{},
			conns: map[string]ConnectionInterface{
				"test-query": &Connection{
					session: mockSession,
				},
				"another-test-query": &Connection{
					session: mockSession,
				},
			},
			crs: &check.CheckResultSet{
				Check: check.NewCheck([]byte(`{
	  "id":"chPzAHTTP",
	  "zone_id":"pzA",
	  "details":{"url":"localhost"},
	  "type":"remote.http",
	  "timeout":15,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"",
	  "disabled":false
				}`)),
			},
			expectedErr: false,
		},
		{
			name:   "No connections",
			ctx:    context.Background(),
			config: &config.Config{},
			conns:  map[string]ConnectionInterface{},
			crs: &check.CheckResultSet{
				Check: check.NewCheck([]byte(`{
	  "id":"chPzAHTTP",
	  "zone_id":"pzA",
	  "details":{"url":"localhost"},
	  "type":"remote.http",
	  "timeout":15,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"",
	  "disabled":false
				}`)),
			},
			expectedErr: false,
		},
		{
			name:        "Connections set to nil",
			ctx:         context.Background(),
			config:      &config.Config{},
			conns:       nil,
			expectedErr: true,
			crs: &check.CheckResultSet{
				Check: check.NewCheck([]byte(`{
	  "id":"chPzAHTTP",
	  "zone_id":"pzA",
	  "details":{"url":"localhost"},
	  "type":"remote.http",
	  "timeout":15,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"",
	  "disabled":false
				}`)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &ConnectionStream{
				ctx:       tt.ctx,
				stopCh:    tt.stopCh,
				config:    tt.config,
				connsMu:   tt.connsMu,
				conns:     tt.conns,
				wg:        tt.wg,
				scheduler: tt.scheduler,
			}
			if tt.expectedErr {
				log.Println("aqui")
				//mockSession.EXPECT().Send(gomock.Any()).Times(0)
				assert.Error(t, cs.SendMetrics(tt.crs))
			} else {
				// at most send 1 request
				mockSession.EXPECT().Send(gomock.Any()).Times(int(math.Min(float64(len(tt.conns)), 1)))
				assert.NoError(t, cs.SendMetrics(tt.crs))
			}
			// wait for 25 milliseconds for the test to run through everything
			// WARNING: this could be flaky; however, it passed 100 iterations locally
			// this is something to watch
			time.Sleep(25 * time.Millisecond)
		})
	}
}
