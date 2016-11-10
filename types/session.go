package types

import (
	"context"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"io"
	"sync"
	"time"
	"github.com/racker/rackspace-monitoring-poller/hostinfo"
)

type CompletionFrame struct {
	Id     uint64
	Method string
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

	sendCh chan Frame

	heartbeatInterval time.Duration
}

func newSession(ctx context.Context, connection *Connection) *Session {
	session := &Session{
		connection:        connection,
		enc:               json.NewEncoder(connection.conn),
		dec:               json.NewDecoder(connection.conn),
		seq:               1,
		sendCh:            make(chan Frame, 128),
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
	s.Send(NewHandshakeRequest(cfg))
}

func (s *Session) Send(msg Frame) {
	msg.SetId(s)
	msg.SetTarget("endpoint")
	msg.SetSource(s.connection.guid)
	s.sendCh <- msg
}

func (s *Session) SetHeartbeatInterval(timeout uint64) {
	duration := time.Duration(timeout) * time.Millisecond
	log.Debugf("setting heartbeat interval %v", duration)
	s.heartbeatInterval = time.Duration(duration)
}

func (s *Session) getCompletionRequest(resp Frame) *CompletionFrame {
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

func (s *Session) handleResponse(resp *FrameMsg) {
	if req := s.getCompletionRequest(resp); req != nil {
		switch req.Method {
		case "handshake.hello":
			resp := NewHandshakeResponse(resp)
			s.SetHeartbeatInterval(resp.Result.HandshakeInterval)
			s.Send(NewPollerRegister([]string{"pzA"}))
		case "check_schedule.get":
		case "poller.register":
		case "heartbeat.post":
		case "check_metrics.post":
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
			f := new(FrameMsg)
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
func (s *Session) handleFrame(f *FrameMsg) {
	js, _ := f.Encode()
	log.Debugf("RECV: %s", js)
	switch f.GetMethod() {
	case "": // Responses do not have a method name
		s.handleResponse(f)
	case "poller.checks.add":
		s.connection.GetStream().GetScheduler().Input() <- f
	case "host_info.get":
		go s.handleHostInfo(f)
	case "poller.checks.end":
	default:
		log.Errorf("  Need to handle method: %v", f.GetMethod())
	}
}

func (s *Session) handleHostInfo(f *FrameMsg) {
	if hinfo := hostinfo.NewHostInfo(*f.GetRawParams()); hinfo != nil {
		go func(s *Session, hinfo hostinfo.HostInfo, f *FrameMsg) {
			cr, err := hinfo.Run()
			if err != nil {
			} else {
				response := NewHostInfoResponse(cr, f, hinfo)
				s.Send(response)
			}
		}(s, hinfo, f)
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
			s.Send(NewHeartbeat())
		}
	}
done:
	log.Debug("heartbeat exiting")
}

func (s *Session) addCompletion(frame Frame) {
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
	log.Println(err)
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
		return
	}
	s.shutdown = true
	s.shutdownLock.Unlock()
	s.cancel()
}

func (s *Session) Wait() {
	<-s.ctx.Done()
}
