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
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/mock_golang"
	"github.com/racker/rackspace-monitoring-poller/poller"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type frameMatcher struct {
	expectedMethod string
}

func frameBytes(expectedMethod string) gomock.Matcher {
	return frameMatcher{expectedMethod: expectedMethod}
}

func (fm frameMatcher) Matches(in interface{}) bool {
	frameBytes, ok := in.([]byte)
	if !ok {
		fmt.Println("not bytes")
		return false
	}

	var frameMsg protocol.FrameMsg
	err := json.Unmarshal(frameBytes, &frameMsg)
	if err != nil {
		fmt.Println("unable to unmarshal", err)
		return false
	}

	if fm.expectedMethod != "" {
		if frameMsg.Method != fm.expectedMethod {
			fmt.Println("wrong method, got", frameMsg.Method)
			return false
		}
	}

	return true
}

func (fm frameMatcher) String() string {
	return fmt.Sprintf("is a frame expecting method '%s'", fm.expectedMethod)
}

func TestEleSession_HeartbeatSending(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	eleConn := poller.NewMockConnection(ctrl)

	conn := mock_golang.NewMockConn(ctrl)

	connStream := poller.NewMockConnectionStream(ctrl)

	config := &config.Config{}
	connStream.EXPECT().GetConfig().AnyTimes().Return(config)

	eleConn.EXPECT().GetConnection().AnyTimes().Return(conn)
	eleConn.EXPECT().GetStream().AnyTimes().Return(connStream)
	eleConn.EXPECT().GetGUID().AnyTimes().Return("1-2-3")
	eleConn.EXPECT().SetReadDeadline(gomock.Any())
	eleConn.EXPECT().SetWriteDeadline(gomock.Any()).AnyTimes()

	newline := []byte{'\r', '\n'}

	conn.EXPECT().Read(gomock.Any()).AnyTimes()

	handshake := conn.EXPECT().Write(frameBytes(protocol.MethodHandshakeHello))
	handshakeNL := conn.EXPECT().Write(newline).After(handshake)

	heartbeat := conn.EXPECT().Write(frameBytes(protocol.MethodHeartbeatPost)).After(handshakeNL).Do(func(content []byte) {
		var parsed protocol.HeartbeatRequest
		err := json.Unmarshal(content, &parsed)
		assert.NoError(t, err)

		require.NotEqual(t, 0, parsed.Params.Timestamp)
	})
	conn.EXPECT().Write(newline).After(heartbeat)

	ctx := context.Background()
	es := poller.NewSession(ctx, eleConn)
	es.SetHeartbeatInterval(10) // resets the heartbeat waiting

	time.Sleep(20 * time.Millisecond) // ...so give it a little longer than that

	es.Close() // to "stop" its go routines
}
