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

func setupConnStreamExpectations(ctrl *gomock.Controller) (eleConn *poller.MockConnection,
	reconciler *poller.MockChecksReconciler,
	writesHere *utils.BlockingReadBuffer, readsHere *utils.BlockingReadBuffer) {
	eleConn = poller.NewMockConnection(ctrl)

	writesHere = utils.NewBlockingReadBuffer()
	readsHere = utils.NewBlockingReadBuffer()

	connStream := poller.NewMockConnectionStream(ctrl)

	eleConn.EXPECT().GetFarendWriter().AnyTimes().Return(writesHere)
	eleConn.EXPECT().GetFarendReader().AnyTimes().Return(readsHere)
	eleConn.EXPECT().GetStream().AnyTimes().Return(connStream)
	eleConn.EXPECT().GetGUID().AnyTimes().Return("1-2-3")
	eleConn.EXPECT().SetReadDeadline(gomock.Any()).AnyTimes()
	eleConn.EXPECT().SetWriteDeadline(gomock.Any()).AnyTimes()

	reconciler = poller.NewMockChecksReconciler(ctrl)

	return
}

func (fm frameMatcher) String() string {
	return fmt.Sprintf("is a frame expecting method '%s'", fm.expectedMethod)
}

func installDeterministicTimestamper(startingTimestamp, timestampInc int64) utils.NowTimestampMillisFunc {
	origTimestamper := utils.NowTimestampMillis
	mockTimestamp := startingTimestamp
	utils.NowTimestampMillis = func() int64 {
		ts := mockTimestamp
		mockTimestamp += timestampInc
		return ts
	}

	return origTimestamper
}

func prepareHandshakeResponse(heartbeatInterval uint64, readsHere io.Writer) {
	var handshakeResp protocol.HandshakeResponse
	handshakeResp.Id = 1                                       // since that's what poller will send as first message
	handshakeResp.Result.HeartbeatInterval = heartbeatInterval // ms
	json.NewEncoder(readsHere).Encode(handshakeResp)
}

func TestEleSession_HeartbeatSending(t *testing.T) {
	if testing.Verbose() {
		logrus.SetLevel(logrus.DebugLevel)
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	eleConn, reconciler, writesHere, readsHere := setupConnStreamExpectations(ctrl)
	defer readsHere.Close()

	origTimestamper := installDeterministicTimestamper(1000, 2000)
	defer func() { utils.NowTimestampMillis = origTimestamper }()

	const heartbeatInterval = 10 // ms

	es := poller.NewSession(context.Background(), eleConn, reconciler, &config.Config{})
	defer es.Close()

	// decoder is used to consume frames sent out by the poller under test
	decoder := json.NewDecoder(writesHere)

	// We should see a handshake, but can ignore it
	handshakeReq := new(protocol.HandshakeRequest)
	err := decoder.Decode(handshakeReq)
	require.NoError(t, err)

	prepareHandshakeResponse(heartbeatInterval, readsHere)

	// Within 2.5 scaled time, we should see two heartbeats and then nothing ready yet
	time.Sleep((heartbeatInterval * 2.5) * time.Millisecond)
	heartbeatReq := new(protocol.HeartbeatRequest)

	// 1 heartbeat
	err = decoder.Decode(heartbeatReq)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), heartbeatReq.Params.Timestamp, "wrong timestamp")

	// ...2 heartbeats
	err = decoder.Decode(heartbeatReq)
	require.NoError(t, err)
	assert.Equal(t, int64(3000), heartbeatReq.Params.Timestamp, "wrong 2nd timestamp")

	assert.False(t, writesHere.ReadReady(), "buffer should have been empty")
}

func TestEleSession_HeartbeatConsumption(t *testing.T) {
	if testing.Verbose() {
		logrus.SetLevel(logrus.DebugLevel)
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	eleConn, reconciler, writesHere, readsHere := setupConnStreamExpectations(ctrl)
	defer readsHere.Close()

	origTimestamper := installDeterministicTimestamper(1000, 2000)
	defer func() { utils.NowTimestampMillis = origTimestamper }()

	const heartbeatInterval = 10 // ms

	es := poller.NewSession(context.Background(), eleConn, reconciler, &config.Config{})
	defer es.Close()

	// allow for handshake resp to fire up heartbeating
	time.Sleep((heartbeatInterval * 1.5) * time.Millisecond)

	// decoder is used to consume frames sent out by the poller under test
	decoder := json.NewDecoder(writesHere)

	// We should see a handshake, but can ignore it
	handshakeReq := new(protocol.HandshakeRequest)
	err := decoder.Decode(handshakeReq)
	require.NoError(t, err)

	prepareHandshakeResponse(heartbeatInterval, readsHere)

	heartbeatReq := new(protocol.HeartbeatRequest)
	err = decoder.Decode(heartbeatReq)
	// sanity check
	assert.Equal(t, int64(1000), heartbeatReq.Params.Timestamp, "wrong timestamp")
	require.NoError(t, err)

	heartbeatResp := new(protocol.HeartbeatResponse)
	heartbeatResp.Id = heartbeatReq.Id
	// simulate a delay of about 1000ms each way and 500ms clock offset
	heartbeatResp.Result.Timestamp = 2500
	json.NewEncoder(readsHere).Encode(heartbeatResp)

	time.Sleep((heartbeatInterval * 0.1) * time.Millisecond)
	offset := es.GetClockOffset()
	latency := es.GetLatency()

	assert.Equal(t, int64(500), offset, "wrong offset")
	assert.Equal(t, int64(2000), latency, "wrong latency")
}
