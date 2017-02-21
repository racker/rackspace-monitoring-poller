package poller_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/poller"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"github.com/stretchr/testify/assert"
	"sync"
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
				conn.EXPECT().Close()
			}

			connFactory := func(address string, guid string, stream poller.ChecksReconciler) poller.Connection {
				return conn
			}

			cs := poller.NewCustomConnectionStream(&config.Config{
				UseSrv:     tt.useSrv,
				Addresses:  tt.addresses(),
				SrvQueries: tt.serverQueries(),
			}, nil, connFactory)
			cs.Connect()

			mockConnWaiting.Wait()
			cs.Stop()

			assert.NotEmpty(t, cs.StopNotify())
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

				c2 := poller.NewMockConnection(ctrl)
				c2.EXPECT().GetLatency().Return(int64(20))

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

func TestEleConnectionStream_SendMetrics_Normal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var cfg config.Config
	connectionFactory := func(address string, guid string, checksReconciler poller.ChecksReconciler) poller.Connection {
		return nil
	}

	cs := poller.NewCustomConnectionStream(&cfg, nil, connectionFactory)

	c1 := poller.NewMockConnection(ctrl)
	c1.EXPECT().GetLatency().AnyTimes().Return(int64(50))

	mockSession := poller.NewMockSession(ctrl)
	mockSession.EXPECT().Send(gomock.Any())
	var msgId uint64 = 20
	mockSession.EXPECT().AssignFrameId(gomock.Any()).AnyTimes().Do(func(msg protocol.Frame) {
		msg.SetId(&msgId)
	})

	c2 := poller.NewMockConnection(ctrl)
	c2.EXPECT().GetLatency().AnyTimes().Return(int64(10))
	c2.EXPECT().GetClockOffset().AnyTimes().Return(int64(0))
	c2.EXPECT().GetSession().AnyTimes().Return(mockSession)

	c3 := poller.NewMockConnection(ctrl)
	c3.EXPECT().GetLatency().AnyTimes().Return(int64(20))

	cs.RegisterConnection("h1", c1)
	cs.RegisterConnection("h2", c2)
	cs.RegisterConnection("h3", c3)

	crs := check.ResultSet{
		Check: &check.TCPCheck{},
	}
	err := cs.SendMetrics(&crs)
	assert.NoError(t, err)

}

func TestEleConnectionStream_SendMetrics_NoConnections(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var cfg config.Config
	connectionFactory := func(address string, guid string, checksReconciler poller.ChecksReconciler) poller.Connection {
		return nil
	}

	cs := poller.NewCustomConnectionStream(&cfg, nil, connectionFactory)

	crs := check.ResultSet{
		Check: &check.TCPCheck{},
	}
	err := cs.SendMetrics(&crs)
	assert.EqualError(t, err, poller.ErrNoConnections.Error())

}
