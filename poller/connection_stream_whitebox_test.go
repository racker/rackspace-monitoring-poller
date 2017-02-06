package poller

import (
	"sync"
	"testing"

	"context"

	"time"

	"math"

	"github.com/golang/mock/gomock"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectionStream_Stop(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConn := NewMockConnection(mockCtrl)

	tests := []struct {
		name        string
		ctx         context.Context
		stopCh      chan struct{}
		config      *config.Config
		connsMu     sync.Mutex
		conns       map[string]Connection
		wg          sync.WaitGroup
		schedulers  map[string]Scheduler
		expectedErr bool
	}{
		{
			name:   "Happy path - one connection",
			ctx:    context.Background(),
			stopCh: make(chan struct{}, 1),
			config: &config.Config{},
			conns: map[string]Connection{
				"test-query": mockConn,
			},
			expectedErr: false,
		},
		{
			name:   "Happy path - two connections",
			ctx:    context.Background(),
			stopCh: make(chan struct{}, 1),
			config: &config.Config{},
			conns: map[string]Connection{
				"test-query":         mockConn,
				"another-test-query": mockConn,
			},
			expectedErr: false,
		},
		{
			name:        "No connections",
			ctx:         context.Background(),
			stopCh:      make(chan struct{}, 1),
			config:      &config.Config{},
			conns:       map[string]Connection{},
			expectedErr: false,
		},
		{
			name:        "Connections set to nil",
			ctx:         context.Background(),
			stopCh:      make(chan struct{}, 1),
			config:      &config.Config{},
			conns:       nil,
			expectedErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &EleConnectionStream{
				ctx:        tt.ctx,
				stopCh:     tt.stopCh,
				config:     tt.config,
				connsMu:    tt.connsMu,
				conns:      tt.conns,
				wg:         tt.wg,
				schedulers: tt.schedulers,
			}
			if tt.expectedErr {
				mockConn.EXPECT().Close().Times(0)
			} else {
				mockConn.EXPECT().Close().Times(len(tt.conns))
			}

			timer := time.NewTimer(25 * time.Millisecond)

			select {
			case <-timer.C:
				go cs.Stop()
			}
		})
	}
}

func TestConnection_StopNotify(t *testing.T) {
	cs := NewConnectionStream(config.NewConfig("test-guid", false), nil)
	assert.Equal(t, cs.(*EleConnectionStream).stopCh, cs.StopNotify())
}

func TestConnection_WaitCh(t *testing.T) {
	cs := NewConnectionStream(config.NewConfig("test-guid", false), nil)
	result := utils.Timebox(t, 25*time.Millisecond, func(t *testing.T) {
		<-cs.WaitCh()
	})
	assert.True(t, result, "wait channel never notified")
}

func TestConnectionStream_SendMetrics(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockSession := NewMockSession(mockCtrl)
	mockSession.EXPECT().GetLatency().AnyTimes().Return(int64(0))
	mockSession.EXPECT().GetClockOffset().AnyTimes().Return(int64(0))
	ctx := context.Background()

	basicCheck, err := check.NewCheck(ctx, []byte(`{
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
				}`))
	require.NoError(t, err)

	tests := []struct {
		name        string
		stopCh      chan struct{}
		config      *config.Config
		connsMu     sync.Mutex
		conns       map[string]Connection
		wg          sync.WaitGroup
		schedulers  map[string]Scheduler
		crs         *check.ResultSet
		expectedErr bool
	}{
		{
			name:   "Happy path - one session",
			config: &config.Config{},
			conns: map[string]Connection{
				"test-query": &EleConnection{
					session: mockSession,
				},
			},
			crs: &check.ResultSet{
				Check: basicCheck,
			},
			expectedErr: false,
		},
		{
			name:   "Happy path - two connections",
			config: &config.Config{},
			conns: map[string]Connection{
				"test-query": &EleConnection{
					session: mockSession,
				},
				"another-test-query": &EleConnection{
					session: mockSession,
				},
			},
			crs: &check.ResultSet{
				Check: basicCheck,
			},
			expectedErr: false,
		},
		{
			name:   "No connections",
			config: &config.Config{},
			conns:  map[string]Connection{},
			crs: &check.ResultSet{
				Check: basicCheck,
			},
			expectedErr: true,
		},
		{
			name:        "Connections set to nil",
			config:      &config.Config{},
			conns:       nil,
			expectedErr: true,
			crs: &check.ResultSet{
				Check: basicCheck,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &EleConnectionStream{
				ctx:        ctx,
				stopCh:     tt.stopCh,
				config:     tt.config,
				connsMu:    tt.connsMu,
				conns:      tt.conns,
				wg:         tt.wg,
				schedulers: tt.schedulers,
			}
			if tt.expectedErr {
				//mockSession.EXPECT().Send(gomock.Any()).Times(0)
				assert.Error(t, cs.SendMetrics(tt.crs))
			} else {
				// at most send 1 request
				mockSession.EXPECT().Send(gomock.Any()).Times(int(math.Min(float64(len(tt.conns)), 1)))
				assert.NoError(t, cs.SendMetrics(tt.crs))
			}
		})
	}
}
