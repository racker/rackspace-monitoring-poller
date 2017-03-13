package poller_test

import (
	"testing"

	"container/list"
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/poller"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"time"
)

func TestConnectionStream_Connect(t *testing.T) {

	tests := []struct {
		name                    string
		addresses               func() []string
		serverQueries           func() []string
		useSrv                  bool
		neverAttemptsConnection bool
	}{
		{
			name: "Happy path",
			addresses: func() []string {
				return []string{"localhost"}
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
			useSrv:                  true,
			neverAttemptsConnection: true,
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var mockConnWaiting sync.WaitGroup

			conn := poller.NewMockConnection(ctrl)
			if !tt.neverAttemptsConnection {
				mockConnWaiting.Add(1)
				connectCall := conn.EXPECT().Connect(gomock.Any(), gomock.Any(), gomock.Any())
				conn.EXPECT().Wait().After(connectCall).Do(func() {
					mockConnWaiting.Done()
				})
				conn.EXPECT().GetLogPrefix().Return("1234")
				conn.EXPECT().Close()
			}

			connFactory := func(address string, guid string, stream poller.ChecksReconciler) poller.Connection {
				return conn
			}

			ctx, cancel := context.WithCancel(context.Background())
			cs := poller.NewCustomConnectionStream(ctx, &config.Config{
				UseSrv:     tt.useSrv,
				Addresses:  tt.addresses(),
				SrvQueries: tt.serverQueries(),
			}, nil, connFactory, nil)
			cs.Connect()

			mockConnWaiting.Wait()
			cancel()

			assert.NotEmpty(t, cs.Wait())
		})
	}
}

func TestConnectionsByHost_ChooseBest(t *testing.T) {

	tests := []struct {
		name      string
		fill      func(poller.ConnectionsByHost, *gomock.Controller)
		expectKey string
	}{
		{
			name: "multi",
			fill: func(conns poller.ConnectionsByHost, ctrl *gomock.Controller) {
				c1 := poller.NewMockConnection(ctrl)
				c1.EXPECT().GetLatency().Return(int64(50))
				c1.EXPECT().GetLogPrefix().AnyTimes().Return("1-2-3")

				c2 := poller.NewMockConnection(ctrl)
				c2.EXPECT().GetLatency().Return(int64(20))
				c2.EXPECT().GetLogPrefix().AnyTimes().Return("1-2-3")

				conns["h1"] = c1
				conns["h2"] = c2
			},
			expectKey: "h2",
		},
		{
			name: "single",
			fill: func(conns poller.ConnectionsByHost, ctrl *gomock.Controller) {
				c1 := poller.NewMockConnection(ctrl)
				c1.EXPECT().GetLatency().Return(int64(50))

				conns["h1"] = c1
			},
			expectKey: "h1",
		},
		{
			name: "empty",
			fill: func(conns poller.ConnectionsByHost, ctrl *gomock.Controller) {
			},
			expectKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var conns poller.ConnectionsByHost = make(poller.ConnectionsByHost)

			tt.fill(conns, ctrl)

			result := conns.ChooseBest()

			if tt.expectKey != "" {
				assert.Equal(t, conns[tt.expectKey], result)
			} else {
				assert.Nil(t, result)
			}

		})
	}
}

type mockConnFactory struct {
	conns *list.List
	used  chan struct{}
}

func newMockConnFactory() *mockConnFactory {
	return &mockConnFactory{
		conns: list.New(),
		used:  make(chan struct{}, 1),
	}
}

func (f *mockConnFactory) add(conn *poller.MockConnection) *poller.MockConnection {
	f.conns.PushBack(conn)
	return conn
}

func (f *mockConnFactory) produce(address string, guid string, checksReconciler poller.ChecksReconciler) poller.Connection {
	if f.conns.Len() > 0 {
		connection := f.conns.Remove(f.conns.Front()).(poller.Connection)
		if f.conns.Len() == 0 {
			close(f.used)
		}
		return connection
	} else {
		return nil
	}
}

func (f *mockConnFactory) wait() error {
	select {
	case <-f.used:
		return nil
	case <-time.After(5 * time.Millisecond):
		return errors.New("Didn't consume all expected connections")
	}
}

func (f *mockConnFactory) renderConfig() *config.Config {
	cfg := &config.Config{
		UseSrv:    false,
		Addresses: make([]string, f.conns.Len()),
	}

	for i := 0; i < f.conns.Len(); i++ {
		cfg.Addresses[i] = fmt.Sprintf("c%d", i+1)
	}

	return cfg
}

func TestEleConnectionStream_SendMetrics_Normal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	factory := newMockConnFactory()
	c1 := factory.add(poller.NewMockConnection(ctrl))
	c2 := factory.add(poller.NewMockConnection(ctrl))
	c3 := factory.add(poller.NewMockConnection(ctrl))

	cfg := factory.renderConfig()

	failedMetrics := make(chan *check.ResultSet, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cs := poller.NewCustomConnectionStream(ctx, cfg, nil, factory.produce, func(crs *check.ResultSet) {
		failedMetrics <- crs
	})
	cs.Connect()
	require.NoError(t, factory.wait())

	c1.EXPECT().GetLatency().AnyTimes().Return(int64(50))
	c1.EXPECT().GetLogPrefix().AnyTimes().Return("1-2-3")

	mockSession := poller.NewMockSession(ctrl)
	mockSession.EXPECT().Send(gomock.Any())

	c2.EXPECT().GetLatency().AnyTimes().Return(int64(10))
	c2.EXPECT().GetClockOffset().AnyTimes().Return(int64(0))
	c2.EXPECT().GetSession().Return(mockSession)
	c2.EXPECT().GetLogPrefix().AnyTimes().Return("1-2-3")

	c3.EXPECT().GetLatency().AnyTimes().Return(int64(20))
	c3.EXPECT().GetLogPrefix().AnyTimes().Return("1-2-3")

	crs := check.ResultSet{
		Check: &check.TCPCheck{},
	}
	cs.SendMetrics(&crs)

	// allow for send metrics channel
	time.Sleep(5 * time.Millisecond)

	assert.Empty(t, failedMetrics)
}

func TestEleConnectionStream_SendMetrics_OneThenDrop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c2Done := make(chan struct{}, 1)
	factory := newMockConnFactory()
	c1 := factory.add(poller.NewMockConnection(ctrl))
	c2 := factory.add(poller.NewMockConnection(ctrl))
	c2.EXPECT().Wait().AnyTimes().Return(c2Done)

	cfg := factory.renderConfig()

	failedMetrics := make(chan *check.ResultSet, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cs := poller.NewCustomConnectionStream(ctx, cfg, nil, factory.produce, func(crs *check.ResultSet) {
		failedMetrics <- crs
	})
	cs.Connect()
	require.NoError(t, factory.wait())

	mock2Session := poller.NewMockSession(ctrl)
	mock2Session.EXPECT().Send(gomock.Any()).Do(func(protocol.Frame) {
		// let one post through, then "close" the connection
		close(c2Done)
	})

	c1.EXPECT().GetLatency().AnyTimes().Return(int64(50))
	c1.EXPECT().GetLogPrefix().AnyTimes().Return("1-2-3")

	c2.EXPECT().GetLatency().AnyTimes().Return(int64(10))
	c2.EXPECT().GetClockOffset().AnyTimes().Return(int64(0))
	c2.EXPECT().GetSession().Return(mock2Session)
	c2.EXPECT().GetLogPrefix().AnyTimes().Return("1-2-3")

	crs := check.ResultSet{
		Check: &check.TCPCheck{},
	}
	cs.SendMetrics(&crs)

	// wait for the "closure" we induced above
	utils.Timebox(t, 5*time.Millisecond, func(t *testing.T) {
		<-c2Done
	})
	assert.Empty(t, failedMetrics)

	// post again after all connections gone
	cs.SendMetrics(&crs)

}

func TestEleConnectionStream_SendMetrics_NoConnections(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	factory := newMockConnFactory()

	mockConn := factory.add(poller.NewMockConnection(ctrl))
	mockConn.EXPECT().Connect(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("simulating failed connect"))

	failedMetrics := make(chan *check.ResultSet, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cs := poller.NewCustomConnectionStream(ctx, factory.renderConfig(), nil, factory.produce, func(crs *check.ResultSet) {
		failedMetrics <- crs
	})
	cs.Connect()
	err := factory.wait()
	require.NoError(t, err)

	crs := &check.ResultSet{
		Check: &check.TCPCheck{},
	}
	cs.SendMetrics(crs)

	select {
	case m := <-failedMetrics:
		assert.Equal(t, crs, m)
	case <-time.After(5 * time.Millisecond):
		assert.Fail(t, "Expected a failed metric")
	}

}
