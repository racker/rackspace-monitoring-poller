package poller_test

import (
	"testing"

	"context"
	"github.com/golang/mock/gomock"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/poller"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"os"
	"strconv"
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
			conn.EXPECT().Connect(gomock.Any(), gomock.Any(), gomock.Not(nil))
			conn.EXPECT().Done().AnyTimes().Return(done)
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

func TestEleConnectionStream_ReConnectOldOnes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	connClosed := make(chan struct{}, 1)
	reconnect := make(chan struct{}, 1)
	closureTimeChan := make(chan time.Time, 1)

	conn := NewMockConnection(ctrl)
	conn.EXPECT().Connect(gomock.Any(), gomock.Any(), gomock.Not(nil))
	conn.EXPECT().Done().AnyTimes().Return(connClosed)
	conn.EXPECT().GetLogPrefix().AnyTimes().Return("1234")
	conn.EXPECT().Close().AnyTimes().Do(func() {
		t.Log("Mock conn is closing")
		close(connClosed)
		closureTimeChan <- time.Now()
	})
	preAuthedChannel := make(chan struct{}, 1)
	close(preAuthedChannel)
	conn.EXPECT().Authenticated().AnyTimes().Return(preAuthedChannel)

	connFactory := func(address string, guid string, stream poller.ChecksReconciler) poller.Connection {
		// only provide mock connection until first closure
		select {
		case <-connClosed:
			reconnect <- struct{}{}
			return nil
		default:
			return conn
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cs := poller.NewCustomConnectionStream(ctx, &config.Config{
		MaxConnectionAge:       100 * time.Millisecond,
		MaxConnectionAgeJitter: 10 * time.Millisecond,
		ReconnectMinBackoff:    1 * time.Millisecond,
		ReconnectMaxBackoff:    1 * time.Millisecond,
		UseSrv:                 false,
		Addresses:              []string{"localhost"},
		SrvQueries:             nil,
	}, nil, connFactory)

	consumer := newPhasingEventConsumer()
	cs.RegisterEventConsumer(consumer)

	connectTime := time.Now()
	cs.Connect()
	consumer.waitFor(t, 15*time.Millisecond, poller.EventTypeRegister, gomock.Eq(conn))

	select {
	case closureTime := <-closureTimeChan:
		closureDelta := closureTime.Sub(connectTime)
		// expected closure upper bound is at least max age + max age jitter + 10ms more for timing fuzziness
		assert.True(t, closureDelta >= 100*time.Millisecond && closureDelta < 120*time.Millisecond,
			"Closure delta is out of bounds", closureDelta)
		break
	case <-time.After(150 * time.Millisecond):
		assert.Fail(t, "Didn't see connection stream closure")
	}

	select {
	case <-reconnect:
		break
	case <-time.After(10 * time.Millisecond):
		assert.Fail(t, "Didn't see re-connect")
	}
}

func TestEleConnectionStream_Connect_ClearText(t *testing.T) {
	os.Setenv(config.EnvCleartext, config.EnabledEnvOpt)
	defer os.Unsetenv(config.EnvCleartext)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	conn := NewMockConnection(ctrl)
	// NOTE: expecting the nil tlsConfig is the primary expectation of this case
	conn.EXPECT().Connect(gomock.Any(), gomock.Any(), gomock.Nil())
	conn.EXPECT().GetLogPrefix().AnyTimes().Return("1234")

	preAuthedChannel := make(chan struct{}, 1)
	close(preAuthedChannel)
	conn.EXPECT().Authenticated().AnyTimes().Return(preAuthedChannel)

	done := make(chan struct{}, 1)
	conn.EXPECT().Done().AnyTimes().Return(done)
	conn.EXPECT().Close().AnyTimes().Do(func() {
		t.Log("Mock conn is closing")
		close(done)
	})

	connFactory := func(address string, guid string, stream poller.ChecksReconciler) poller.Connection {
		return conn
	}

	ctx, cancel := context.WithCancel(context.Background())
	cs := poller.NewCustomConnectionStream(ctx, &config.Config{
		UseSrv:    false,
		Addresses: []string{"localhost"},
	}, nil, connFactory)

	consumer := newPhasingEventConsumer()
	cs.RegisterEventConsumer(consumer)

	cs.Connect()
	consumer.waitFor(t, 10*time.Millisecond, poller.EventTypeRegister, gomock.Eq(conn))
	cancel()
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
				c1.EXPECT().HasLatencyMeasurements().AnyTimes().Return(true)
				c1.EXPECT().GetLatency().Return(int64(50))
				c1.EXPECT().GetLogPrefix().AnyTimes().Return("1-2-3")

				c2 := NewMockConnection(ctrl)
				c2.EXPECT().HasLatencyMeasurements().AnyTimes().Return(true)
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
				c1.EXPECT().HasLatencyMeasurements().AnyTimes().Return(true)
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

	cfg := factory.renderConfig(4)

	consumer := newPhasingEventConsumer()

	ctx, _ := context.WithCancel(context.Background())

	cs := poller.NewCustomConnectionStream(ctx, cfg, nil, factory.produce)
	cs.RegisterEventConsumer(consumer)

	cs.Connect()
	factory.waitForConnections(t, 4, 20*time.Millisecond)

	// None authenticated, so none should be registered yet
	consumer.assertNoEvent(t, 5*time.Millisecond)

	// ...now mark authenticated
	c1.SetAuthenticated()
	c2.SetAuthenticated()
	c3.SetAuthenticated()
	c4.SetAuthenticated()
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeRegister, gomock.Eq(c1))
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeRegister, gomock.Eq(c2))
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeRegister, gomock.Eq(c3))
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeRegister, gomock.Eq(c4))

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
	c1.EXPECT().HasLatencyMeasurements().AnyTimes().Return(true)
	c1.EXPECT().GetLogPrefix().AnyTimes().Return("c1")
	c1.EXPECT().Done().AnyTimes().Return(done)

	c2 := factory.add(NewMockConnection(ctrl))
	c2.EXPECT().GetLatency().AnyTimes().Return(int64(10))
	c2.EXPECT().HasLatencyMeasurements().AnyTimes().Return(true)
	c2.EXPECT().GetClockOffset().AnyTimes().Return(int64(0))
	c2.EXPECT().GetSession().Return(mockSession)
	c2.EXPECT().GetLogPrefix().AnyTimes().Return("c2")
	c2.EXPECT().Done().AnyTimes().Return(doneEarly)

	c3 := factory.add(NewMockConnection(ctrl))
	c3.EXPECT().GetLatency().AnyTimes().Return(int64(20))
	c3.EXPECT().HasLatencyMeasurements().AnyTimes().Return(true)
	c3.EXPECT().GetClockOffset().AnyTimes().Return(int64(0))
	c3.EXPECT().GetLogPrefix().AnyTimes().Return("c3")
	c3.EXPECT().Done().AnyTimes().Return(done)
	c3.EXPECT().GetSession().Return(mockSession)

	cfg := factory.renderConfig(3)

	consumer := newPhasingEventConsumer()

	ctx, _ := context.WithCancel(context.Background())

	cs := poller.NewCustomConnectionStream(ctx, cfg, nil, factory.produce)
	cs.RegisterEventConsumer(consumer)

	cs.Connect()
	factory.waitForConnections(t, 3, 20*time.Millisecond)
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

	c1 := factory.add(NewMockConnection(ctrl))
	c1.EXPECT().GetLatency().AnyTimes().Return(int64(10))
	c1.EXPECT().HasLatencyMeasurements().AnyTimes().Return(true)
	c1.EXPECT().GetClockOffset().AnyTimes().Return(int64(0))
	c1.EXPECT().GetSession().Return(mockSession)
	c1.EXPECT().GetLogPrefix().AnyTimes().Return("c2")
	c1.EXPECT().Done().AnyTimes().Return(done)

	cfg := factory.renderConfig(1)

	consumer := newPhasingEventConsumer()

	ctx, _ := context.WithCancel(context.Background())

	cs := poller.NewCustomConnectionStream(ctx, cfg, nil, factory.produce)
	cs.RegisterEventConsumer(consumer)

	cs.Connect()
	factory.waitForConnections(t, 1, 20*time.Millisecond)
	c1.SetAuthenticated()
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeRegister, gomock.Eq(c1))

	crs := &check.ResultSet{
		Check: &check.TCPCheck{},
	}
	cs.SendMetrics(crs)

	factory.waitForFrame(t, 5*time.Millisecond)
	consumer.assertNoEvent(t, 5*time.Millisecond)

	close(done)
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeDeregister, gomock.Eq(c1))
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeAllConnectionsLost, gomock.Nil())

	cs.SendMetrics(crs)
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeDroppedMetric, gomock.Eq(crs))
}

func setupMockConnExpectations(c *MockConnection, latency int, offset int, mockSession *MockSession, prefix string) chan struct{} {
	done := make(chan struct{}, 3)
	c.EXPECT().GetLatency().AnyTimes().Return(int64(latency))
	c.EXPECT().HasLatencyMeasurements().AnyTimes().Return(true)
	c.EXPECT().GetClockOffset().AnyTimes().Return(int64(offset))
	c.EXPECT().GetLogPrefix().AnyTimes().Return(prefix)
	c.EXPECT().Done().AnyTimes().Return(done)

	return done
}

func TestEleConnectionStream_NewGuidAfterLosingAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	factory := newMockConnFactory()

	mockSession := NewMockSession(ctrl)

	consumer := newPhasingEventConsumer()

	ctx, _ := context.WithCancel(context.Background())

	c1 := factory.add(NewMockConnection(ctrl))
	c1done := setupMockConnExpectations(c1, 10, 0, mockSession, "c1")

	cfg := factory.renderConfig(1)
	cfg.Guid = "original-guid"
	cfg.ReconnectMinBackoff = 1 * time.Millisecond

	cs := poller.NewCustomConnectionStream(ctx, cfg, nil, factory.produce)
	cs.RegisterEventConsumer(consumer)

	cs.Connect()
	factory.waitForConnections(t, 1, 20*time.Millisecond)
	c1.SetAuthenticated()
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeRegister, gomock.Eq(c1))

	assert.Equal(t, "original-guid", factory.guids[0])

	c2 := factory.add(NewMockConnection(ctrl))
	c2done := setupMockConnExpectations(c2, 10, 0, mockSession, "c2")

	close(c1done)
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeDeregister, gomock.Eq(c1))
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeAllConnectionsLost, gomock.Nil())

	factory.waitForConnections(t, 1, 20*time.Millisecond)
	c2.SetAuthenticated()
	consumer.waitFor(t, 5*time.Millisecond, poller.EventTypeRegister, gomock.Eq(c2))

	assert.NotEqual(t, factory.guids[0], factory.guids[1])

	close(c2done)
}

func TestEleConnectionStream_SendMetrics_NoConnections(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	factory := newMockConnFactory()

	cfg := factory.renderConfig(1)

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

func TestEleConnectionStream_UseStatsdDistributor(t *testing.T) {
	factory := newMockConnFactory()

	udpConn, err := net.ListenUDP("udp4", nil)
	require.NoError(t, err)
	defer udpConn.Close()

	pktChan := make(chan string, 100)
	done := make(chan struct{}, 1)
	defer close(done)
	go func() {
		for {
			select {
			case <-done:
				return

			default:
				buf := make([]byte, 500)
				len, _, err := udpConn.ReadFrom(buf)
				if err != nil {
					t.Log(err)
				} else {
					pktChan <- string(buf[:len])
				}
			}
		}
	}()

	cfg := factory.renderConfig(1)
	cfg.Token = "***.123456"
	addr := udpConn.LocalAddr()
	ourUdpAddr := addr.(*net.UDPAddr)
	cfg.StatsdEndpoint = net.JoinHostPort("127.0.0.1", strconv.FormatInt(int64(ourUdpAddr.Port), 10))

	consumer := newPhasingEventConsumer()

	ctx, _ := context.WithCancel(context.Background())

	cs := poller.NewCustomConnectionStream(ctx, cfg, nil, factory.produce)
	cs.RegisterEventConsumer(consumer)
	cs.Connect()

	crs := check.ResultSet{
		Check: &check.PingCheck{},
	}
	crs.Check.SetCheckType("remote.ping")
	crs.Check.SetID("chTEST")

	metrics := make(map[string]*metric.Metric)
	metrics["average"] = metric.NewMetric("average", "", metric.MetricFloat, 4.5, "")
	crs.Metrics = []*check.Result{
		{Metrics: metrics},
	}

	cs.SendMetrics(&crs)

	select {
	case <-time.After(1 * time.Second):
		t.Fail()

	case pkt := <-pktChan:
		assert.Regexp(t, "rackspace.remote.ping.average:4.500000|g|#tenant:123456,entity:,check:chTEST,targetIP:,host:.+", pkt)
	}
}
