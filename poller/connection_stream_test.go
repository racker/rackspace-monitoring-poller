package poller_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/poller"
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
