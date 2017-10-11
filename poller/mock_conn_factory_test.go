//
// Copyright 2017 Rackspace
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS-IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package poller_test

import (
	"container/list"
	"github.com/golang/mock/gomock"
	"crypto/tls"
	"testing"
	"github.com/racker/rackspace-monitoring-poller/poller"
	"fmt"
	"github.com/stretchr/testify/require"
	"time"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/stretchr/testify/assert"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"context"
)

type mockConnFactory struct {
	conns     *list.List
	connected chan struct{}
	frames    chan protocol.Frame
	guids     []string
}

func newMockConnFactory() *mockConnFactory {
	return &mockConnFactory{
		conns:     list.New(),
		connected: make(chan struct{}, 10),
		frames:    make(chan protocol.Frame, 10),
		guids:     make([]string, 0),
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
	conn.EXPECT().SetAuthenticated().Do(func() {
		close(auth)
	})
	return conn
}

func (f *mockConnFactory) interceptSend(frame protocol.Frame) {
	f.frames <- frame
}

func (f *mockConnFactory) waitForFrame(t *testing.T, timeout time.Duration) protocol.Frame {
	select {
	case <-time.After(timeout):
		assert.Fail(t, "Did not see frame")
		return nil
	case f := <-f.frames:
		return f
	}
}

func (f *mockConnFactory) produce(address string, guid string, checksReconciler poller.ChecksReconciler) poller.Connection {
	if f.conns.Len() > 0 {
		connection := f.conns.Remove(f.conns.Front()).(poller.Connection)
		f.guids = append(f.guids, guid)
		return connection
	} else {
		return nil
	}
}

func (f *mockConnFactory) waitForConnections(t *testing.T, expect int, timeout time.Duration) {
	var seen int

	limiter := time.After(timeout)

	for {
		select {
		case <-limiter:
			require.Fail(t, "Did not see enough connections")
			return

		case <-f.connected:
			seen++
			if seen >= expect {
				return
			}
		}
	}
}

func (f *mockConnFactory) renderConfig(addresses int) *config.Config {
	cfg := &config.Config{
		UseSrv:    false,
		Addresses: make([]string, addresses),
	}

	for i := 0; i < addresses; i++ {
		cfg.Addresses[i] = fmt.Sprintf("c%d", i+1)
	}

	return cfg
}
