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
		scheduler   map[string]*Scheduler
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
	assert.Equal(t, cs.(*ConnectionStream).stopCh, cs.StopNotify())
}

func TestConnection_WaitCh(t *testing.T) {
	cs := NewConnectionStream(config.NewConfig("test-guid", false), nil)
	result := Timebox(t, 25*time.Millisecond, func(t *testing.T) {
		<-cs.WaitCh()
	})
	assert.True(t, result, "wait channel never notified")
}

func TestConnection_GetScheduler(t *testing.T) {
	cs := NewConnectionStream(config.NewConfig("test-guid", false), nil)
	assert.Equal(t, cs.(*ConnectionStream).scheduler, cs.GetScheduler())
}

func TestConnectionStream_SendMetrics(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockSession := NewMockSessionInterface(mockCtrl)
	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	tests := []struct {
		name        string
		ctx         context.Context
		stopCh      chan struct{}
		config      *config.Config
		connsMu     sync.Mutex
		conns       map[string]ConnectionInterface
		wg          sync.WaitGroup
		scheduler   map[string]*Scheduler
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
				}`), cancelCtx, cancelFunc),
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
				}`), cancelCtx, cancelFunc),
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
				}`), cancelCtx, cancelFunc),
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
				}`), cancelCtx, cancelFunc),
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

// Timebox is used for putting a time bounds around a chunk of code, given as the function boxed.
// NOTE that if the duration d elapses, then boxed will be left to run off in its go-routine...it can't be
// forcefully terminated.
// This function can be used outside of a unit test context by passing nil for t
// Returns true if boxed finished before duration d elapsed.
func Timebox(t *testing.T, d time.Duration, boxed func(t *testing.T)) bool {
	timer := time.NewTimer(d)
	completed := make(chan struct{})

	go func() {
		boxed(t)
		close(completed)
	}()

	select {
	case <-timer.C:
		if t != nil {
			t.Fatal("Timebox expired")
		}
		return false
	case <-completed:
		timer.Stop()
		return true
	}
}
