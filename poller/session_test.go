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
	"github.com/Sirupsen/logrus"
	"github.com/golang/mock/gomock"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/mock_golang"
	"github.com/racker/rackspace-monitoring-poller/poller"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
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

func setupConnStreamExpectations(ctrl *gomock.Controller) (eleConn *poller.MockConnection, conn *mock_golang.MockConn, connStream *poller.MockConnectionStream) {
	eleConn = poller.NewMockConnection(ctrl)

	conn = mock_golang.NewMockConn(ctrl)

	connStream = poller.NewMockConnectionStream(ctrl)

	config := &config.Config{}
	connStream.EXPECT().GetConfig().AnyTimes().Return(config)

	eleConn.EXPECT().GetConnection().AnyTimes().Return(conn)
	eleConn.EXPECT().GetStream().AnyTimes().Return(connStream)
	eleConn.EXPECT().GetGUID().AnyTimes().Return("1-2-3")
	eleConn.EXPECT().SetReadDeadline(gomock.Any())
	eleConn.EXPECT().SetWriteDeadline(gomock.Any()).AnyTimes()

	return
}

func (fm frameMatcher) String() string {
	return fmt.Sprintf("is a frame expecting method '%s'", fm.expectedMethod)
}

func TestEleSession_HeartbeatSending(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	eleConn, conn, _ := setupConnStreamExpectations(ctrl)

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

func TestEleSession_HeartbeatConsumption(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	newline := []byte{'\r', '\n'}

	eleConn, conn, _ := setupConnStreamExpectations(ctrl)

	responseCh := make(chan []byte, 1)
	serverCh := make(chan []byte, 2)

	firstRead := conn.EXPECT().Read(gomock.Any()).Do(func(readBuffer []byte) (n int, err error) {
		response := <-responseCh
		n = copy(readBuffer, response)
		err = nil

		t.Log("Populated read buffer", n, string(readBuffer[:n]))
		return
	})
	conn.EXPECT().Read(gomock.Any()).AnyTimes().After(firstRead).Do(func(readBuffer []byte) (n int, err error) {
		return 0, io.EOF
	})

	origTimestamper := utils.NowTimestampMillis
	mockTimestamp := int64(1000)
	utils.NowTimestampMillis = func() int64 {
		ts := mockTimestamp
		mockTimestamp += 2000
		return ts
	}

	// semi-mock server
	go func() {

		heartbeatBytes := <-serverCh
		<-serverCh // newline

		t.Log("Processing heartbeat bytes")
		f := new(protocol.FrameMsg)
		err := json.Unmarshal(heartbeatBytes, f)
		require.NoError(t, err)
		t.Log("Got hearbeat frame", f)

		req := protocol.HeartbeatRequest{FrameMsg: *f}
		err = json.Unmarshal(f.RawParams, &req.Params)
		require.NoError(t, err)
		t.Log("Got heartbeat req", req)

		resp := protocol.DecodeHeartbeatResponse(f)
		resp.Method = protocol.MethodEmpty
		resp.RawParams = nil
		resp.Result.Timestamp = 1500

		respBytes, err := json.Marshal(resp)
		require.NoError(t, err)

		respBytes = append(respBytes, '\r', '\n')

		t.Log("Providing response bytes", len(respBytes), string(respBytes))
		responseCh <- respBytes
	}()

	handshake := conn.EXPECT().Write(frameBytes(protocol.MethodHandshakeHello))
	handshakeNL := conn.EXPECT().Write(newline).After(handshake)

	heartbeat := conn.EXPECT().Write(frameBytes(protocol.MethodHeartbeatPost)).After(handshakeNL).Do(func(content []byte) {
		t.Log("DBG saw write heartbeat post", content)
		serverCh <- content
	})
	lastWrite := conn.EXPECT().Write(newline).After(heartbeat).Do(func(content []byte) {
		t.Log("DBG saw write newline", content)
		serverCh <- content
	})
	conn.EXPECT().Write(gomock.Any()).AnyTimes().After(lastWrite)

	ctx := context.Background()
	es := poller.NewSession(ctx, eleConn)
	es.SetHeartbeatInterval(10) // resets the heartbeat waiting

	time.Sleep(500 * time.Millisecond) // ...so give it a little longer than that

	latency := es.GetLatency()
	offset := es.GetClockOffset()

	assert.Equal(t, int64(1000), latency)
	assert.Equal(t, int64(500), offset)

	utils.NowTimestampMillis = origTimestamper
	es.Close() // to "stop" its go routines

}
