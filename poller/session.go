//
// Copyright 2016 Rackspace
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

package poller

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"github.com/racker/rackspace-monitoring-poller/utils"
)

// CompletionFrame is a pointer to a request with a specified
// method used for the request
type CompletionFrame struct {
	ID     uint64
	Method string
}

const (
	sendChannelSize = 128
)

// EleSession implements Session interface
// See Session for more information
type EleSession struct {
	// reference to the connection
	connection       Connection
	checksReconciler ChecksReconciler

	config *config.Config

	// Used to cancel all go routines
	ctx    context.Context
	cancel context.CancelFunc

	// JSON encoders and decoder streams
	enc *json.Encoder
	dec *json.Decoder

	// sequence message ID
	seq uint64

	shutdownLock sync.Mutex
	shutdown     bool
	error        error

	completionsMu sync.Mutex
	completions   map[uint64]*CompletionFrame

	sendCh chan protocol.Frame

	heartbeatsStarter    sync.Once
	heartbeatInterval    time.Duration
	heartbeatResponses   chan *protocol.HeartbeatResponse
	heartbeatMeasurement struct {
		tracking      utils.TimeLatencyTracking
		expectedSeqID uint64
		offset        int64
		latency       int64
	}
}

func NewSession(ctx context.Context, connection Connection, checksReconciler ChecksReconciler, config *config.Config) Session {
	session := &EleSession{
		connection:         connection,
		checksReconciler:   checksReconciler,
		config:             config,
		dec:                json.NewDecoder(connection.GetFarendReader()),
		seq:                0, // so that handshake req gets ID 1 after incrementing
		sendCh:             make(chan protocol.Frame, sendChannelSize),
		heartbeatResponses: make(chan *protocol.HeartbeatResponse, 1),
		completions:        make(map[uint64]*CompletionFrame),
	}
	ctx, cancel := context.WithCancel(ctx)
	go session.runFrameReading(ctx)
	go session.runFrameSending(ctx)
	session.ctx = ctx
	session.cancel = cancel
	session.Auth()
	return session
}

// Auth sends a handshake request with token, agent id, name,
// and process version
func (s *EleSession) Auth() {
	request := protocol.NewHandshakeRequest(s.config)
	request.SetId(&s.seq)
	s.Send(request)
}

// Send stages a frame for sending after setting the target and source.
// NOTE: The protocol.Frame.SetId MUST be called prior to this method.
func (s *EleSession) Send(msg protocol.Frame) {
	msg.SetTarget("endpoint")
	msg.SetSource(s.connection.GetGUID())
	s.sendCh <- msg
}

// Respond is equivalent to Send but improves readability by emphasizing this is the poller responding to a
// request from the server.
func (s *EleSession) Respond(msg protocol.Frame) {
	msg.SetTarget("endpoint")
	msg.SetSource(s.connection.GetGUID())
	s.sendCh <- msg
}

func (s *EleSession) getCompletionRequest(resp protocol.Frame) *CompletionFrame {
	s.completionsMu.Lock()
	req, ok := s.completions[resp.GetId()]
	if !ok {
		s.completionsMu.Unlock()
		return nil
	}
	delete(s.completions, resp.GetId())
	s.completionsMu.Unlock()
	return req
}

func (s *EleSession) handleResponse(resp *protocol.FrameMsg) {
	if req := s.getCompletionRequest(resp); req != nil {
		switch req.Method {
		case protocol.MethodHandshakeHello:
			resp := protocol.DecodeHandshakeResponse(resp)
			s.heartbeatInterval = time.Duration(resp.Result.HeartbeatInterval) * time.Millisecond

			// just to be sure guard against multiple handshake starting multiple heartbeat routines
			s.heartbeatsStarter.Do(s.goRunHeartbeats)
		case protocol.MethodHeartbeatPost:
			resp := protocol.DecodeHeartbeatResponse(resp)
			s.heartbeatResponses <- resp
		case protocol.MethodCheckScheduleGet:
		case protocol.MethodPollerRegister:
		case protocol.MethodCheckMetricsPost:
		default:
			log.Errorf("Unexpected method: %s", req.Method)
		}
	}
}

// GetReadDeadline adds sessions's heartbeat interval to configured read deadline
func (s *EleSession) computeReadDeadline() time.Time {
	return s.config.ComputeReadDeadline(s.heartbeatInterval)
}

// GetWriteDeadline adds sessions's heartbeat interval to configured write deadline
func (s *EleSession) computeWriteDeadline() time.Time {
	return s.config.ComputeWriteDeadline(s.heartbeatInterval)
}

func (s *EleSession) runFrameReading(ctx context.Context) {
	log.Debug("read starting")
	for {
		select {
		case <-ctx.Done():
			goto done
		default:
			f := new(protocol.FrameMsg)
			s.connection.SetReadDeadline(s.computeReadDeadline())
			if err := s.dec.Decode(f); err == io.EOF {
				goto done
			} else if err != nil {
				s.exitError(err)
				goto done
			}
			go s.handleFrame(f)
		}
	}
done:
	log.Debug("read exiting")
	s.cancel()
}

func (s *EleSession) handleFrame(f *protocol.FrameMsg) {
	js, _ := f.Encode()
	if log.GetLevel() >= log.DebugLevel {
		log.Debugf("RECV: %s", js)
	}
	switch f.GetMethod() {
	case protocol.MethodEmpty: // Responses do not have a method name
		s.handleResponse(f)
	case protocol.MethodPollerChecksAdd:
		// in process of being modified
		// s.connection.GetStream().GetScheduler().Input() <- f
	case protocol.MethodHostInfoGet:
		go s.handleHostInfo(f)
	case protocol.MethodPollerChecksEnd:
	default:
		log.Errorf("  Need to handle method: %v", f.GetMethod())
	}
}

func (s *EleSession) handleHostInfo(f *protocol.FrameMsg) {
	if hinfo := hostinfo.NewHostInfo(f.GetRawParams()); hinfo != nil {
		cr, err := hinfo.Run()
		if err != nil {
		} else {
			response := hostinfo.NewHostInfoResponse(cr, f, hinfo)
			s.Respond(response)
		}
	}
}

// runHeartbeats is driven by a timer on heartbeatInterval
func (s *EleSession) runHeartbeats() {
	log.Debug("heartbeat starting")
	for {
		select {
		case <-s.ctx.Done():
			goto done
		case <-time.After(s.heartbeatInterval):
			req := protocol.NewHeartbeatRequest()
			req.SetId(&s.seq)

			s.prepareHeartbeatMeasurement(req)
			log.WithField("req", req).Debug("Initiating heartbeat")
			s.Send(req)

		case resp := <-s.heartbeatResponses:
			s.updateHeartbeatMeasurement(resp)

		}
	}
done:
	log.Debug("heartbeat exiting")
}

func (s *EleSession) goRunHeartbeats() {
	go s.runHeartbeats()
}

func (s *EleSession) prepareHeartbeatMeasurement(req *protocol.HeartbeatRequest) {
	s.heartbeatMeasurement.expectedSeqID = req.Id
	s.heartbeatMeasurement.tracking.PollerSendTimestamp = req.Params.Timestamp
}

func (s *EleSession) updateHeartbeatMeasurement(resp *protocol.HeartbeatResponse) {
	if s.heartbeatMeasurement.expectedSeqID != resp.Id {
		log.WithFields(log.Fields{
			"expected": s.heartbeatMeasurement.expectedSeqID,
			"received": resp.Id,
		}).Warn("Received out of sequence heartbeat response. Unable to compute latency from it.")
		return
	}

	s.heartbeatMeasurement.tracking.PollerRecvTimestamp = utils.NowTimestampMillis()
	s.heartbeatMeasurement.tracking.ServerRecvTimestamp = resp.Result.Timestamp
	s.heartbeatMeasurement.tracking.ServerRespTimestamp = resp.Result.Timestamp

	offset, latency, err := s.heartbeatMeasurement.tracking.ComputeSkew()
	if err != nil {
		log.WithField("err", err).Warn("Failed to compute skew")
		return
	}

	log.WithFields(log.Fields{
		"offset": offset,
		"delay":  latency,
	}).Debug("Computed poller-server latencies")

	s.heartbeatMeasurement.offset = offset
	s.heartbeatMeasurement.latency = latency
}

func (s *EleSession) GetClockOffset() int64 {
	return s.heartbeatMeasurement.offset
}

func (s *EleSession) GetLatency() int64 {
	return s.heartbeatMeasurement.latency
}

func (s *EleSession) addCompletion(frame protocol.Frame) {
	cFrame := &CompletionFrame{ID: frame.GetId(), Method: frame.GetMethod()}
	s.completionsMu.Lock()
	defer s.completionsMu.Unlock()
	s.completions[cFrame.ID] = cFrame
}

// runs it it's own go routine and consumes from sendCh
func (s *EleSession) runFrameSending(ctx context.Context) {
	log.Debug("send starting")
	for {
		select {
		case <-ctx.Done():
			goto done
		case frame := <-s.sendCh:
			s.addCompletion(frame)
			s.connection.SetWriteDeadline(s.computeWriteDeadline())
			data, err := frame.Encode()
			if err != nil {
				s.exitError(err)
				goto done
			}
			if log.GetLevel() >= log.DebugLevel {
				log.Debugf("SEND: %s", data)
			}
			_, err = s.connection.GetFarendWriter().Write(data)
			if err != nil {
				s.exitError(err)
				goto done
			}
			_, err = s.connection.GetFarendWriter().Write([]byte{'\r', '\n'})
			if err != nil {
				s.exitError(err)
				goto done
			}
		}
	}
done:
	log.Debug("send exiting")
	s.Close()
}

func (s *EleSession) exitError(err error) {
	log.Warn("Session exiting with error", err)
	s.shutdownLock.Lock()
	if s.error == nil {
		s.error = err
	}
	s.shutdownLock.Unlock()
	s.Close()
}

// Close shuts down session's context and closes session
func (s *EleSession) Close() {
	s.shutdownLock.Lock()
	if s.shutdown {
		s.shutdownLock.Unlock()
		return
	}
	s.shutdown = true
	s.shutdownLock.Unlock()
	s.cancel()
}

// Wait waits for the context to complete
func (s *EleSession) Wait() {
	<-s.ctx.Done()
}
