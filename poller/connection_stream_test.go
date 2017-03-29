package poller_test

import (
	"testing"

	"container/list"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/poller"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/stretchr/testify/assert"
	"time"
)

func TestConnectionStream_Connect(t *testing.T) {

	tests := []struct {
		name                string
		addresses           func() []string
		serverQueries       func() []string
		useSrv              bool
		connectionTimeoutMs int
	}{
		{
			name: "Happy path",
			addresses: func() []string {
				return []string{"localhost"}
			},
			serverQueries: func() []string {
				return []string{}
			},
			useSrv:              false,
			connectionTimeoutMs: 15,
		},
		{
			name: "Use service",
			addresses: func() []string {
				return []string{}
			},
			serverQueries: func() []string {
				return []string{"_monitoringagent._tcp.dfw1.prod.monitoring.api.rackspacecloud.com"}
			},
			useSrv:              true,
			connectionTimeoutMs: 500,
		},
		{
			name: "Invalid url",
			addresses: func() []string {
				return []string{"invalid-url:1234"}
			},
			serverQueries: func() []string {
				return []string{}
			},
			useSrv:              false,
			connectionTimeoutMs: 15,
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			done := make(chan struct{}, 1)

			conn := NewMockConnection(ctrl)
			conn.EXPECT().Connect(gomock.Any(), gomock.Any(), gomock.Any())
			conn.EXPECT().Done().Return(done)
			conn.EXPECT().GetLogPrefix().AnyTimes().Return("1234")
			conn.EXPECT().Close().AnyTimes().Do(func() {
				t.Log("Mock conn is closing")
				close(done)
			})
			preAuthedChannel := make(chan struct{}, 1)
			close(preAuthedChannel)
			conn.EXPECT().Authenticated().AnyTimes().Return(preAuthedChannel)

			connFactory := func(address string, guid string, stream poller.ChecksReconciler) poller.Connection {
				return conn
			}

			ctx, cancel := context.WithCancel(context.Background())
			cs := poller.NewCustomConnectionStream(ctx, &config.Config{
				UseSrv:     tt.useSrv,
				Addresses:  tt.addresses(),
				SrvQueries: tt.serverQueries(),
			}, nil, connFactory)

			consumer := newPhasingEventConsumer()
			cs.RegisterEventConsumer(consumer)

			cs.Connect()
			consumer.waitFor(t, time.Duration(tt.connectionTimeoutMs)*time.Millisecond, poller.EventTypeRegister, gomock.Eq(conn))
			cancel()

			select {
			case <-cs.Done():
				break
			case <-time.After(5 * time.Millisecond):
				assert.Fail(t, "Didn't see connection stream closure")
			}
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
				c1 := NewMockConnection(ctrl)
				c1.EXPECT().GetLatency().Return(int64(50))
				c1.EXPECT().GetLogPrefix().AnyTimes().Return("1-2-3")

				c2 := NewMockConnection(ctrl)
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
				c1 := NewMockConnection(ctrl)
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

func TestEleConnectionStream_SendMetrics_Normal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	done := make(chan struct{}, 3)
	factory := newMockConnFactory()

	mockSession := NewMockSession(ctrl)
	mockSession.EXPECT().Send(gomock.Any()).Do(factory.interceptSend)

	c1 := factory.add(NewMockConnection(ctrl))
	c1.EXPECT().GetLatency().AnyTimes().Return(int64(50))
	c1.EXPECT().HasLatencyMeasurements().AnyTimes().Return(true)
	c1.EXPECT().GetLogPrefix().AnyTimes().Return("c1")
	c1.EXPECT().Done().AnyTimes().Return(done)

	c2 := factory.add(NewMockConnection(ctrl))
	c2.EXPECT().GetLatency().AnyTimes().Return(int64(10))
	c2.EXPECT().HasLatencyMeasurements().AnyTimes().Return(true)
	c2.EXPECT().GetClockOffset().AnyTimes().Return(int64(0))
	c2.EXPECT().GetSession().Return(mockSession)
	c2.EXPECT().GetLogPrefix().AnyTimes().Return("c2")
	c2.EXPECT().Done().AnyTimes().Return(done)

	c3 := factory.add(NewMockConnection(ctrl))
	c3.EXPECT().GetLatency().AnyTimes().Return(int64(20))
	c3.EXPECT().HasLatencyMeasurements().AnyTimes().Return(true)
	c3.EXPECT().GetLogPrefix().AnyTimes().Return("c3")
	c3.EXPECT().Done().AnyTimes().Return(done)

	c4 := factory.add(NewMockConnection(ctrl))
	c4.EXPECT().GetLatency().AnyTimes().Return(int64(0))
	c4.EXPECT().HasLatencyMeasurements().AnyTimes().Return(false)
	c4.EXPECT().GetLogPrefix().AnyTimes().Return("c4")
	c4.EXPECT().Done().AnyTimes().Return(done)

	cfg := factory.renderConfig()

	consumer := newPhasingEventConsumer()

	ctx, _ := context.WithCancel(context.Background())

	cs := poller.NewCustomConnectionStream(ctx, cfg, nil, factory.produce)
	cs.RegisterEventConsumer(consumer)

	cs.Connect()
	factory.waitForConnections(t, 20*time.Millisecond)

	// None authenticated, so none should be registered yet
	consumer.assertNoEvent(t, 5*time.Millisecond)

	// ...now mark authenticated
	c1.SetAuthenticated()
	c2.SetAuthenticated()
	c3.SetAuthenticated()
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeRegister, gomock.Eq(c1))
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeRegister, gomock.Eq(c2))
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeRegister, gomock.Eq(c3))

	crs := check.ResultSet{
		Check: &check.TCPCheck{},
	}
	cs.SendMetrics(&crs)

	factory.waitForFrame(t, 5*time.Millisecond)
	consumer.assertNoEvent(t, 5*time.Millisecond)
}

func TestEleConnectionStream_SendMetrics_RollOver(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	done := make(chan struct{}, 3)
	doneEarly := make(chan struct{}, 3)
	factory := newMockConnFactory()

	// this session gets shared between c2 and c3
	mockSession := NewMockSession(ctrl)
	mockSession.EXPECT().Send(gomock.Any()).Times(2).Do(factory.interceptSend)

	c1 := factory.add(NewMockConnection(ctrl))
	c1.EXPECT().GetLatency().AnyTimes().Return(int64(50))
	c1.EXPECT().GetLogPrefix().AnyTimes().Return("c1")
	c1.EXPECT().Done().AnyTimes().Return(done)

	c2 := factory.add(NewMockConnection(ctrl))
	c2.EXPECT().GetLatency().AnyTimes().Return(int64(10))
	c2.EXPECT().GetClockOffset().AnyTimes().Return(int64(0))
	c2.EXPECT().GetSession().Return(mockSession)
	c2.EXPECT().GetLogPrefix().AnyTimes().Return("c2")
	c2.EXPECT().Done().AnyTimes().Return(doneEarly)

	c3 := factory.add(NewMockConnection(ctrl))
	c3.EXPECT().GetLatency().AnyTimes().Return(int64(20))
	c3.EXPECT().GetClockOffset().AnyTimes().Return(int64(0))
	c3.EXPECT().GetLogPrefix().AnyTimes().Return("c3")
	c3.EXPECT().Done().AnyTimes().Return(done)
	c3.EXPECT().GetSession().Return(mockSession)

	cfg := factory.renderConfig()

	consumer := newPhasingEventConsumer()

	ctx, _ := context.WithCancel(context.Background())

	cs := poller.NewCustomConnectionStream(ctx, cfg, nil, factory.produce)
	cs.RegisterEventConsumer(consumer)

	cs.Connect()
	factory.waitForConnections(t, 20*time.Millisecond)
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeRegister, gomock.Eq(c1))
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeRegister, gomock.Eq(c2))
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeRegister, gomock.Eq(c3))

	crs := check.ResultSet{
		Check: &check.TCPCheck{},
	}
	cs.SendMetrics(&crs)

	factory.waitForFrame(t, 5*time.Millisecond)
	consumer.assertNoEvent(t, 5*time.Millisecond)

	close(doneEarly) // closes c2
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeDeregister, gomock.Eq(c2))

	cs.SendMetrics(&crs)
	factory.waitForFrame(t, 5*time.Millisecond)
	consumer.assertNoEvent(t, 5*time.Millisecond)
}

func TestEleConnectionStream_SendMetrics_OneThenDrop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	done := make(chan struct{}, 3)
	factory := newMockConnFactory()

	mockSession := NewMockSession(ctrl)
	mockSession.EXPECT().Send(gomock.Any()).Do(factory.interceptSend)

	c2 := factory.add(NewMockConnection(ctrl))
	c2.EXPECT().GetLatency().AnyTimes().Return(int64(10))
	c2.EXPECT().GetClockOffset().AnyTimes().Return(int64(0))
	c2.EXPECT().GetSession().Return(mockSession)
	c2.EXPECT().GetLogPrefix().AnyTimes().Return("c2")
	c2.EXPECT().Done().AnyTimes().Return(done)

	cfg := factory.renderConfig()

	consumer := newPhasingEventConsumer()

	ctx, _ := context.WithCancel(context.Background())

	cs := poller.NewCustomConnectionStream(ctx, cfg, nil, factory.produce)
	cs.RegisterEventConsumer(consumer)

	cs.Connect()
	factory.waitForConnections(t, 20*time.Millisecond)
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeRegister, gomock.Eq(c2))

	crs := &check.ResultSet{
		Check: &check.TCPCheck{},
	}
	cs.SendMetrics(crs)

	factory.waitForFrame(t, 5*time.Millisecond)
	consumer.assertNoEvent(t, 5*time.Millisecond)

	close(done)
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeDeregister, gomock.Eq(c2))

	cs.SendMetrics(crs)
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeDroppedMetric, gomock.Eq(crs))
}

func TestEleConnectionStream_SendMetrics_NoConnections(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	factory := newMockConnFactory()

	cfg := factory.renderConfig()

	consumer := newPhasingEventConsumer()

	ctx, _ := context.WithCancel(context.Background())

	cs := poller.NewCustomConnectionStream(ctx, cfg, nil, factory.produce)
	cs.RegisterEventConsumer(consumer)

	cs.Connect()
	consumer.assertNoEvent(t, 5*time.Millisecond)

	crs := &check.ResultSet{
		Check: &check.TCPCheck{},
	}
	cs.SendMetrics(crs)

	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeDroppedMetric, gomock.Eq(crs))
}

type mockConnFactory struct {
	conns     *list.List
	connected chan struct{}
	frames    chan protocol.Frame
}

func newMockConnFactory() *mockConnFactory {
	return &mockConnFactory{
		conns:     list.New(),
		connected: make(chan struct{}, 10),
		frames:    make(chan protocol.Frame, 10),
	}
}

func (f *mockConnFactory) add(conn *MockConnection) *MockConnection {
	f.conns.PushBack(conn)
	auth := make(chan struct{}, 1)
	conn.EXPECT().Connect(gomock.Any(), gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, config *config.Config, tlsConfig *tls.Config) {
			f.connected <- struct{}{}
		})
	conn.EXPECT().Authenticated().AnyTimes().Return(auth)
	conn.EXPECT().SetAuthenticated().AnyTimes().Do(func() {
		close(auth)
	})
	return conn
}

func (f *mockConnFactory) interceptSend(frame protocol.Frame) {
	f.frames <- frame
}

func (f *mockConnFactory) waitForFrame(t *testing.T, timeout time.Duration) {
	select {
	case <-time.After(timeout):
		assert.Fail(t, "Did not see frame")
	case <-f.frames:
		break
	}
}

func (f *mockConnFactory) produce(address string, guid string, checksReconciler poller.ChecksReconciler) poller.Connection {
	if f.conns.Len() > 0 {
		connection := f.conns.Remove(f.conns.Front()).(poller.Connection)
		return connection
	} else {
		return nil
	}
}

func (f *mockConnFactory) waitForConnections(t *testing.T, timeout time.Duration) {
	var seen int

	limiter := time.After(timeout)

	for {
		select {
		case <-limiter:
			assert.Fail(t, "Did not see enough connections. Only saw %v", seen)

		case <-f.connected:
			seen++
			if seen >= f.conns.Len() {
				return
			}
		}
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

type phasingEventConsumer struct {
	events chan utils.Event
}

func newPhasingEventConsumer() *phasingEventConsumer {
	return &phasingEventConsumer{
		events: make(chan utils.Event, 10),
	}
}

func (c *phasingEventConsumer) HandleEvent(evt utils.Event) error {
	c.events <- evt
	return nil
}

func (c *phasingEventConsumer) waitFor(t *testing.T, timeout time.Duration, eventType string, targetMatcher gomock.Matcher) {
	select {
	case evt := <-c.events:
		assert.Equal(t, eventType, evt.Type(), "Wrong event type")
		if !targetMatcher.Matches(evt.Target()) {
			assert.Fail(t, targetMatcher.String())
		}
	case <-time.After(timeout):
		assert.Fail(t, "Did not observe an event")
	}
}

func (c *phasingEventConsumer) assertNoEvent(t *testing.T, timeout time.Duration) {
	select {
	case evt := <-c.events:
		assert.Fail(t, "Should not have seen an event, but got %v", evt)
	case <-time.After(timeout):
		break
	}
}
