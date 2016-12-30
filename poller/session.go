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
	"github.com/racker/rackspace-monitoring-poller/hostinfo"
	"github.com/racker/rackspace-monitoring-poller/protocol"
)

type CompletionFrame struct {
	Id     uint64
	Method string
}

type SessionInterface interface {
	Auth()
	Send(msg protocol.Frame)
	SendResponse(msg protocol.Frame)
	SetHeartbeatInterval(timeout uint64)
	GetReadDeadline() time.Time
	GetWriteDeadline() time.Time
	Close()
	Wait()
}

type Session struct {
	// reference to the connection
	connection *Connection

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

	heartbeatInterval time.Duration
}

func newSession(ctx context.Context, connection *Connection) *Session {
	session := &Session{
		connection:        connection,
		enc:               json.NewEncoder(connection.conn),
		dec:               json.NewDecoder(connection.conn),
		seq:               1,
		sendCh:            make(chan protocol.Frame, 128),
		heartbeatInterval: time.Duration(40 * time.Second),
		completions:       make(map[uint64]*CompletionFrame),
	}
	ctx, cancel := context.WithCancel(ctx)
	go session.read(ctx)
	go session.send(ctx)
	go session.heartbeat(ctx)
	session.ctx = ctx
	session.cancel = cancel
	session.Auth()
	return session
}

func (s *Session) Auth() {
	cfg := s.connection.GetStream().GetConfig()
	s.Send(protocol.NewHandshakeRequest(cfg))
}

func (s *Session) Send(msg protocol.Frame) {
	msg.SetId(&s.seq)
	msg.SetTarget("endpoint")
	msg.SetSource(s.connection.guid)
	s.sendCh <- msg
}

func (s *Session) SendResponse(msg protocol.Frame) {
	msg.SetTarget("endpoint")
	msg.SetSource(s.connection.guid)
	s.sendCh <- msg
}

func (s *Session) SetHeartbeatInterval(timeout uint64) {
	duration := time.Duration(timeout) * time.Millisecond
	log.Debugf("setting heartbeat interval %v", duration)
	s.heartbeatInterval = time.Duration(duration)
}

func (s *Session) getCompletionRequest(resp protocol.Frame) *CompletionFrame {
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

func (s *Session) handleResponse(resp *protocol.FrameMsg) {
	if req := s.getCompletionRequest(resp); req != nil {
		switch req.Method {
		case protocol.MethodHandshakeHello:
			resp := protocol.NewHandshakeResponse(resp)
			s.SetHeartbeatInterval(resp.Result.HandshakeInterval)
		case protocol.MethodCheckScheduleGet:
		case protocol.MethodPollerRegister:
		case protocol.MethodHeartbeatPost:
		case protocol.MethodCheckMetricsPost:
		default:
			log.Errorf("Unexpected method: %s", req.Method)
		}
	}
}

func (s *Session) GetReadDeadline() time.Time {
	return s.connection.GetStream().GetConfig().GetReadDeadline(s.heartbeatInterval)
}

func (s *Session) GetWriteDeadline() time.Time {
	return s.connection.GetStream().GetConfig().GetWriteDeadline(s.heartbeatInterval)
}

// runs it it's own go routine
func (s *Session) read(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			goto done
		default:
			f := new(protocol.FrameMsg)
			s.connection.SetReadDeadline(s.GetReadDeadline())
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

// runs it it's own go routine
func (s *Session) handleFrame(f *protocol.FrameMsg) {
	js, _ := f.Encode()
	log.Debugf("RECV: %s", js)
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

func (s *Session) handleHostInfo(f *protocol.FrameMsg) {
	if hinfo := hostinfo.NewHostInfo(f.GetRawParams()); hinfo != nil {
		cr, err := hinfo.Run()
		if err != nil {
		} else {
			response := hostinfo.NewHostInfoResponse(cr, f, hinfo)
			s.SendResponse(response)
		}
	}
}

// runs it it's own go routine
func (s *Session) heartbeat(ctx context.Context) {
	log.Debug("heartbeat starting")
	for {
		select {
		case <-ctx.Done():
			goto done
		case <-time.After(s.heartbeatInterval):
			s.Send(protocol.NewHeartbeat())
		}
	}
done:
	log.Debug("heartbeat exiting")
}

func (s *Session) addCompletion(frame protocol.Frame) {
	cFrame := &CompletionFrame{Id: frame.GetId(), Method: frame.GetMethod()}
	s.completionsMu.Lock()
	defer s.completionsMu.Unlock()
	s.completions[cFrame.Id] = cFrame
}

// runs it it's own go routine
func (s *Session) send(ctx context.Context) {
	log.Debug("send starting")
	for {
		select {
		case <-ctx.Done():
			goto done
		case frame := <-s.sendCh:
			s.addCompletion(frame)
			s.connection.SetWriteDeadline(s.GetWriteDeadline())
			data, err := frame.Encode()
			if err != nil {
				s.exitError(err)
				goto done
			}
			log.Debugf("SEND: %s", data)
			_, err = s.connection.conn.Write(data)
			if err != nil {
				s.exitError(err)
				goto done
			}
			_, err = s.connection.conn.Write([]byte{'\r', '\n'})
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

func (s *Session) exitError(err error) {
	log.Warn("Session exiting with error", err)
	s.shutdownLock.Lock()
	if s.error == nil {
		s.error = err
	}
	s.shutdownLock.Unlock()
	s.Close()
}

func (s *Session) Close() {
	s.shutdownLock.Lock()
	if s.shutdown {
		s.shutdownLock.Unlock()
		return
	}
	s.shutdown = true
	s.shutdownLock.Unlock()
	s.cancel()
}

func (s *Session) Wait() {
	<-s.ctx.Done()
}
