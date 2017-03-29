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
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/poller"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
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

func setupConnStreamExpectations(ctrl *gomock.Controller, expectAuth bool) (eleConn *MockConnection,
	reconciler *MockChecksReconciler,
	writesHere *utils.BlockingReadBuffer, readsHere *utils.BlockingReadBuffer) {
	eleConn = NewMockConnection(ctrl)

	writesHere = utils.NewBlockingReadBuffer()
	readsHere = utils.NewBlockingReadBuffer()

	eleConn.EXPECT().GetFarendWriter().AnyTimes().Return(writesHere)
	eleConn.EXPECT().GetFarendReader().AnyTimes().Return(readsHere)
	eleConn.EXPECT().GetGUID().AnyTimes().Return("1-2-3")
	eleConn.EXPECT().GetLogPrefix().AnyTimes().Return("1-2-3")
	eleConn.EXPECT().SetReadDeadline(gomock.Any()).AnyTimes()
	eleConn.EXPECT().SetWriteDeadline(gomock.Any()).AnyTimes()
	eleConn.EXPECT().GetLogPrefix().AnyTimes().Return("1-2-3")
	if expectAuth {
		authCh := make(chan struct{}, 1)
		eleConn.EXPECT().SetAuthenticated().Do(func() {
			close(authCh)
		})
		eleConn.EXPECT().Authenticated().AnyTimes().Return(authCh)
	}

	reconciler = NewMockChecksReconciler(ctrl)

	return
}

func (fm frameMatcher) String() string {
	return fmt.Sprintf("is a frame expecting method '%s'", fm.expectedMethod)
}

func installDeterministicTimestamper(startingTimestamp, timestampInc int64) utils.NowTimestampMillisFunc {
	mockTimestamp := startingTimestamp

	return utils.InstallAlternateTimestampFunc(func() int64 {
		ts := mockTimestamp
		mockTimestamp += timestampInc
		return ts
	})
}

func prepareHandshakeResponse(heartbeatInterval uint64, readsHere io.Writer) {
	var handshakeResp protocol.HandshakeResponse
	handshakeResp.Id = 1                                       // since that's what poller will send as first message
	handshakeResp.Result.HeartbeatInterval = heartbeatInterval // ms
	json.NewEncoder(readsHere).Encode(handshakeResp)
}

func prepareHandshakeResponseError(code uint64, message string, readsHere io.Writer) {
	var handshakeResp protocol.HandshakeResponse
	handshakeResp.Id = 1 // since that's what poller will send as first message
	handshakeResp.Error = &protocol.Error{
		Code:    code,
		Message: message,
	}
	json.NewEncoder(readsHere).Encode(handshakeResp)
}

func handshake(t *testing.T, writesHere, readsHere *utils.BlockingReadBuffer, heartbeatInterval uint64) *json.Decoder {
	// decoder is used to consume frames sent out by the poller under test
	decoder := json.NewDecoder(writesHere)
	// We should see a handshake, but can ignore it
	handshakeReq := new(protocol.HandshakeRequest)
	err := decoder.Decode(handshakeReq)
	require.NoError(t, err)
	prepareHandshakeResponse(heartbeatInterval, readsHere)
	return decoder
}

func handshakeError(t *testing.T, writesHere, readsHere *utils.BlockingReadBuffer, code uint64, message string) *json.Decoder {
	// decoder is used to consume frames sent out by the poller under test
	decoder := json.NewDecoder(writesHere)
	// We should see a handshake, but can ignore it
	handshakeReq := &protocol.HandshakeRequest{}
	err := decoder.Decode(handshakeReq)
	require.NoError(t, err)
	prepareHandshakeResponseError(code, message, readsHere)
	return decoder
}

func TestEleSession_HeartbeatSending(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	eleConn, reconciler, writesHere, readsHere := setupConnStreamExpectations(ctrl, true)
	defer readsHere.Close()

	origTimestamper := installDeterministicTimestamper(1000, 2000)
	defer utils.InstallAlternateTimestampFunc(origTimestamper)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	eleConn, reconciler, writesHere, readsHere := setupConnStreamExpectations(ctrl, true)
	defer readsHere.Close()

	origTimestamper := installDeterministicTimestamper(1000, 2000)
	defer utils.InstallAlternateTimestampFunc(origTimestamper)

	const heartbeatInterval = 10 // ms

	es := poller.NewSession(context.Background(), eleConn, reconciler, &config.Config{})
	defer es.Close()

	decoder := handshake(t, writesHere, readsHere, heartbeatInterval)

	// allow for handshake resp to fire up heartbeating
	time.Sleep((heartbeatInterval * 1.5) * time.Millisecond)

	heartbeatReq := new(protocol.HeartbeatRequest)
	err := decoder.Decode(heartbeatReq)
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

func TestEleSession_HandshakeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	eleConn, reconciler, writesHere, readsHere := setupConnStreamExpectations(ctrl, false)
	defer readsHere.Close()

	eventConsumer := newPhasingEventConsumer()
	cfg := config.NewConfig("1-2-3", false, nil)
	es := poller.NewSession(context.Background(), eleConn, reconciler, cfg)
	es.RegisterEventConsumer(eventConsumer)
	defer es.Close()

	handshakeError(t, writesHere, readsHere, 400, "some error")

	eventConsumer.waitFor(t, 10*time.Millisecond, poller.EventTypeReadError, gomock.Any())
}

func TestEleSession_HandshakeTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	eleConn, reconciler, _, readsHere := setupConnStreamExpectations(ctrl, false)
	defer readsHere.Close()

	eleConn.EXPECT().Close()

	eventConsumer := newPhasingEventConsumer()
	cfg := config.NewConfig("1-2-3", false, nil)
	cfg.TimeoutAuth = 10 * time.Millisecond
	es := poller.NewSession(context.Background(), eleConn, reconciler, cfg)
	es.RegisterEventConsumer(eventConsumer)
	defer es.Close()

	eventConsumer.waitFor(t, 2*cfg.TimeoutAuth, poller.EventTypeAuthTimeout, gomock.Any())
}

func TestEleSession_HandshakeTimeoutStoppedOnSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	eleConn, reconciler, writesHere, readsHere := setupConnStreamExpectations(ctrl, true)
	defer readsHere.Close()

	origTimestamper := installDeterministicTimestamper(1000, 2000)
	defer utils.InstallAlternateTimestampFunc(origTimestamper)

	eventConsumer := newPhasingEventConsumer()
	cfg := config.NewConfig("1-2-3", false, nil)
	cfg.TimeoutAuth = 10 * time.Millisecond
	es := poller.NewSession(context.Background(), eleConn, reconciler, cfg)
	es.RegisterEventConsumer(eventConsumer)
	defer es.Close()

	// decoder is used to consume frames sent out by the poller under test
	decoder := json.NewDecoder(writesHere)

	// We should see a handshake, but can ignore it
	handshakeReq := new(protocol.HandshakeRequest)
	err := decoder.Decode(handshakeReq)
	require.NoError(t, err)
	/*
			        soon after this the handshake response is "sent back"
			       /   auth timeout would have fired here, but should be stopped
			      /   /      end of test
			     /   /      /       first heartbeat would have been sent
		      	/   /      /       /
			  0    10     20      30   ms
	*/

	prepareHandshakeResponse(30, readsHere)

	eventConsumer.assertNoEvent(t, 2*cfg.TimeoutAuth)
}

func TestEleSession_PollerPrepare(t *testing.T) {

	tests := []struct {
		name                  string
		prepareSeq            string
		commitSeq             string
		expectedPrepResponses []protocol.PollerPrepareResult

		expectedCommitResponse protocol.PollerCommitResult
		expectValidate         bool
		expectReconcile        bool
		actionCount            int
		expectedZone           string
		reconcileValidateErr   error
	}{
		{
			name:       "normal",
			prepareSeq: "normal",
			commitSeq:  "5",
			expectedPrepResponses: []protocol.PollerPrepareResult{
				{
					Status:  "prepared",
					Version: 5,
				},
			},
			expectedCommitResponse: protocol.PollerCommitResult{
				Status:  "committed",
				Version: 5,
			},
			expectValidate:  true,
			expectReconcile: true,
			actionCount:     1,
			expectedZone:    "zn1",
		},
		{
			name:       "larger",
			prepareSeq: "larger",
			commitSeq:  "",
			expectedPrepResponses: []protocol.PollerPrepareResult{
				{
					Status:  "prepared",
					Version: 618,
				},
			},
			expectValidate: true,
			actionCount:    10,
			expectedZone:   "pzUFXMulHf",
		},
		{
			name:       "oldPrepare",
			prepareSeq: "oldPrepare",
			commitSeq:  "5",
			expectedPrepResponses: []protocol.PollerPrepareResult{
				{
					Status:  "ignored",
					Version: 4,
				},
				{
					Status:  "prepared",
					Version: 5,
				},
			},
			expectedCommitResponse: protocol.PollerCommitResult{
				Status:  "committed",
				Version: 5,
			},
			expectValidate:  true,
			expectReconcile: true,
			actionCount:     1,
			expectedZone:    "zn1",
		},
		{
			name:       "wrongCommit",
			prepareSeq: "normal",
			commitSeq:  "6",
			expectedPrepResponses: []protocol.PollerPrepareResult{
				{
					Status:  "prepared",
					Version: 5,
				},
			},
			expectedCommitResponse: protocol.PollerCommitResult{
				Status:  "ignored",
				Version: 6,
			},
			expectValidate:  true,
			expectReconcile: false,
			actionCount:     1,
			expectedZone:    "zn1",
		},
		{
			name:       "abort",
			prepareSeq: "abort",
			commitSeq:  "",
			expectedPrepResponses: []protocol.PollerPrepareResult{
				{
					Status:  "aborted",
					Version: 5,
				},
			},
			expectValidate:  true,
			expectReconcile: false,
			actionCount:     1,
			expectedZone:    "zn1",
		},
		{
			name:       "badDirective",
			prepareSeq: "badDirective",
			commitSeq:  "",
			expectedPrepResponses: []protocol.PollerPrepareResult{
				{
					Status:  "failed",
					Version: 5,
				},
			},
			expectValidate:  true,
			expectReconcile: false,
			actionCount:     1,
			expectedZone:    "zn1",
		},
		{
			name:       "wrongEndVersion",
			prepareSeq: "wrongEndVersion",
			commitSeq:  "",
			expectedPrepResponses: []protocol.PollerPrepareResult{
				{
					Status:  "failed",
					Version: 6,
				},
			},
			expectValidate:  true,
			expectReconcile: false,
			actionCount:     1,
			expectedZone:    "zn1",
		},
		{
			name:       "missingBlock",
			prepareSeq: "missingBlock",
			commitSeq:  "",
			expectedPrepResponses: []protocol.PollerPrepareResult{
				{
					Status:  "failed",
					Version: 5,
				},
			},
			expectValidate:  true,
			expectReconcile: false,
			actionCount:     1,
			expectedZone:    "zn1",
		},
		{
			name:       "wrongBlockVer",
			prepareSeq: "wrongBlockVer",
			commitSeq:  "",
			expectedPrepResponses: []protocol.PollerPrepareResult{
				{
					Status:  "failed",
					Version: 5,
				},
			},
			expectValidate:  true,
			expectReconcile: false,
			actionCount:     1,
			expectedZone:    "zn1",
		},
		{
			name:       "missingInManifest",
			prepareSeq: "missingInManifest",
			commitSeq:  "",
			expectedPrepResponses: []protocol.PollerPrepareResult{
				{
					Status:  "failed",
					Version: 5,
				},
			},
			expectValidate:  true,
			expectReconcile: false,
			actionCount:     1,
			expectedZone:    "zn1",
		},
		{
			name:       "olderThanCommitted",
			prepareSeq: "olderThanCommitted",
			commitSeq:  "",
			expectedPrepResponses: []protocol.PollerPrepareResult{
				{
					Status:  "ignored",
					Version: -1,
				},
			},
			expectValidate:  false,
			expectReconcile: false,
			actionCount:     1,
			expectedZone:    "zn1",
		},
		{
			name:       "badStartAction",
			prepareSeq: "badStartAction",
			commitSeq:  "",
			expectedPrepResponses: []protocol.PollerPrepareResult{
				{
					Status:  "failed",
					Version: 5,
				},
			},
			expectValidate:  false,
			expectReconcile: false,
			actionCount:     1,
			expectedZone:    "zn1",
		},
		{
			name:       "reconcilerValidateError",
			prepareSeq: "normal",
			commitSeq:  "",
			expectedPrepResponses: []protocol.PollerPrepareResult{
				{
					Status:  "failed",
					Version: 5,
				},
			},
			expectValidate:       true,
			reconcileValidateErr: errors.New("Some kind of inconsistency"),
			expectReconcile:      false,
			actionCount:          1,
			expectedZone:         "zn1",
		},
		{
			name:       "endNoPrep",
			prepareSeq: "endNoPrep",
			commitSeq:  "",
			expectedPrepResponses: []protocol.PollerPrepareResult{
				{
					Status:  "failed",
					Version: 5,
				},
			},
			expectValidate:  false,
			expectReconcile: false,
			actionCount:     1,
			expectedZone:    "zn1",
		},
		{
			name:       "newSupercedesInProgress",
			prepareSeq: "newSupercedesInProgress",
			commitSeq:  "",
			expectedPrepResponses: []protocol.PollerPrepareResult{
				{
					Status:  "ignored",
					Version: 5,
				},
				{
					Status:  "prepared",
					Version: 6,
				},
			},
			expectValidate:  true,
			expectReconcile: false,
			actionCount:     1,
			expectedZone:    "zn1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			eleConn, reconciler, writesHere, readsHere := setupConnStreamExpectations(ctrl, true)
			defer readsHere.Close()

			origTimestamper := installDeterministicTimestamper(1000, 2000)
			defer utils.InstallAlternateTimestampFunc(origTimestamper)

			cfg := config.NewConfig("1-2-3", false, nil)
			es := poller.NewSession(context.Background(), eleConn, reconciler, cfg)
			defer es.Close()

			decoder := handshake(t, writesHere, readsHere, 50000)
			time.Sleep(5 * time.Millisecond)

			if tt.expectValidate {
				reconciler.EXPECT().ValidateChecks(gomock.Any()).Do(func(cp poller.ChecksPreparing) {
					assert.Len(t, cp.GetActionableChecks(), tt.actionCount)
				}).Return(tt.reconcileValidateErr).AnyTimes()
			}

			pollerPrepare, err := ioutil.ReadFile(fmt.Sprintf("testdata/poller_prepare_%s.seq", tt.prepareSeq))
			require.NoError(t, err)
			t.Log("Sending prepare sequence")
			readsHere.Write(pollerPrepare)

			utils.TimeboxNamed(t, tt.name, 10*time.Millisecond, func(t *testing.T) {

				for _, expected := range tt.expectedPrepResponses {
					var resp protocol.PollerPrepareResponse
					decoder.Decode(&resp)
					t.Log("Received prepare response", resp)

					assert.Equal(t, expected.Version, resp.Result.Version)
					assert.Equal(t, expected.Status, resp.Result.Status)
					assert.Equal(t, tt.expectedZone, resp.Result.ZoneId)
				}

			})

			if tt.expectReconcile {
				reconciler.EXPECT().ReconcileChecks(gomock.Any()).Do(func(cp poller.ChecksPrepared) {
					assert.Len(t, cp.GetActionableChecks(), 1)
				})
			}

			if tt.commitSeq != "" {
				pollerCommit, err := ioutil.ReadFile(fmt.Sprintf("testdata/poller_commit_%s.json", tt.commitSeq))
				require.NoError(t, err)
				t.Log("Sending commit", string(pollerCommit))
				readsHere.Write(pollerCommit)
				utils.Timebox(t, 10*time.Millisecond, func(t *testing.T) {

					var resp protocol.PollerPrepareResponse
					decoder.Decode(&resp)
					t.Log("Received commit response", resp)

					assert.Equal(t, tt.expectedCommitResponse.Version, resp.Result.Version)
					assert.Equal(t, tt.expectedCommitResponse.Status, resp.Result.Status)

				})
			}

		})
	}

}

func TestEleSession_PollerPrepareTimeout(t *testing.T) {

	tests := []struct {
		name string
		seq  string
	}{
		{
			name: "hangingPrepare",
			seq:  "hangingPrepare",
		},
		{
			name: "hangingEnd",
			seq:  "hangingEnd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			eleConn, reconciler, writesHere, readsHere := setupConnStreamExpectations(ctrl, true)
			defer readsHere.Close()

			origTimestamper := installDeterministicTimestamper(1000, 2000)
			defer utils.InstallAlternateTimestampFunc(origTimestamper)

			cfg := &config.Config{
				TimeoutPrepareEnd: 10 * time.Millisecond,
			}
			es := poller.NewSession(context.Background(), eleConn, reconciler, cfg)
			defer es.Close()

			decoder := handshake(t, writesHere, readsHere, 50000)
			time.Sleep(5 * time.Millisecond)

			reconciler.EXPECT().ValidateChecks(gomock.Any()).Do(func(cp poller.ChecksPreparing) {
				assert.Len(t, cp.GetActionableChecks(), 1)
			})

			pollerPrepare, err := ioutil.ReadFile(fmt.Sprintf("testdata/poller_prepare_%s.seq", tt.seq))
			require.NoError(t, err)
			t.Log("Sending prepare sequence")
			readsHere.Write(pollerPrepare)

			utils.Timebox(t, 30*time.Millisecond, func(t *testing.T) {

				var resp protocol.PollerPrepareResponse
				decoder.Decode(&resp)
				t.Log("Received prepare response", resp)

				assert.Equal(t, protocol.PrepareResultStatusFailed, resp.Result.Status)

			})

		})
	}

}
