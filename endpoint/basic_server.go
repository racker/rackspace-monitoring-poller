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

package endpoint

import (
	"context"
	"crypto/tls"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"go/types"
	"net"
	"time"
)

const (
	ExpectedAgentHeartbeatSec = 60
	ConnectionWriteAllowance  = 60 * time.Second
)

type BasicServer struct {
	Certificate *tls.Certificate
	BindAddr    string

	*AgentTracker
}

func (s *BasicServer) ApplyConfig(cfg *config.EndpointConfig) error {

	cert, err := LoadCertificateFromConfig(cfg)
	if err != nil {
		return err
	}
	s.Certificate = cert

	bindAddr := cfg.BindAddr
	if bindAddr == "" {
		bindAddr = net.JoinHostPort("", config.DefaultPort)
	}
	s.BindAddr = bindAddr

	s.AgentTracker = NewAgentTracker(cfg)

	return nil
}
func (s *BasicServer) UseMetricsRouter(mr *MetricsRouter) {
	s.AgentTracker.UseMetricsRouter(mr)
}

func (s *BasicServer) ListenAndServe() error {

	rootContext := context.Background()

	s.AgentTracker.Start(rootContext)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*s.Certificate},
	}

	listener, err := tls.Listen("tcp", s.BindAddr, tlsConfig)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.WithField("boundAddr", listener.Addr()).Info("Endpoint is accepting connections from pollers")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Failed to accept connection", err.Error())
		}
		go s.handleConnection(rootContext, conn)
	}
}

func (s *BasicServer) handleConnection(ctx context.Context, c net.Conn) {
	defer c.Close()
	log.WithField("remoteAddr", c.RemoteAddr()).Info("Handling connection")

	smartC := utils.NewSmartConn(c)
	// the agent doesn't know it yet, but we're imposing the keepalive duration upon them to send us a handshake, etc
	smartC.ReadKeepalive = ExpectedAgentHeartbeatSec * time.Second
	smartC.WriteAllowance = ConnectionWriteAllowance
	// this is also where we could setup endpoint->agent heartbeats
	err := smartC.Start()
	if err != nil {
		log.WithField("remoteAddr", c.RemoteAddr()).Errorln("Failed to setup read keepalives", err)
		return
	}

	var frames = make(chan *protocol.FrameMsg, 10)
	go s.frameDecoder(smartC, frames)

	for {
		select {
		case <-ctx.Done():
			log.WithField("remoteAddr", c.RemoteAddr()).Info("Done handling connection due to context being done", ctx.Err())
			return

		case frame := <-frames:
			err := s.consumeFrame(ctx, smartC, frame)
			if err != nil {
				log.Warnln("Failed to consume frame", err)
				// assume the worst, get out, and close the connection
				return
			}

		}

	}
}

func (s *BasicServer) frameDecoder(c *utils.SmartConn, frames chan<- *protocol.FrameMsg) {
	log.WithField("remoteAddr", c.RemoteAddr()).Debug("Frame decoder starting")
	defer log.WithField("remoteAddr", c.RemoteAddr()).Debug("Frame decoder stopped")

	decoder := json.NewDecoder(c)

	for decoder.More() {
		var frame protocol.FrameMsg
		err := decoder.Decode(&frame)
		if err != nil {
			log.WithField("remoteAddr", c.RemoteAddr()).Warn("Failed to decode frame")
			return
		}

		log.WithField("remoteAddr", c.RemoteAddr()).Debug("Received frame")
		frames <- &frame
	}
}

func (s *BasicServer) consumeFrame(ctx context.Context, c *utils.SmartConn, frame *protocol.FrameMsg) error {
	log.WithFields(log.Fields{
		"remoteAddr": c.RemoteAddr(),
		"msgId":      frame.Id,
		"source":     frame.Source,
		"method":     frame.Method,
	}).Debug("Consuming frame")

	switch frame.Method {
	case protocol.MethodHandshakeHello:
		handshakeReq := &protocol.HandshakeRequest{FrameMsg: *frame}
		err := json.Unmarshal(frame.RawParams, &handshakeReq.Params)
		if err != nil {
			logUnmarshalError(frame, c)
			return err
		}

		resp := &protocol.HandshakeResponse{}
		resp.Method = protocol.MethodEmpty
		resp.Id = frame.Id
		resp.Result.HandshakeInterval = ExpectedAgentHeartbeatSec * 1000

		log.Debug("SEND handshake resp", resp)
		c.WriteJSON(resp)

		agentErrors := s.AgentTracker.ProcessHello(*handshakeReq, c)
		go watchForAgentErrors(ctx, agentErrors, c)

	case protocol.MethodPollerRegister:
		pollerRegisterReq := &protocol.PollerRegister{FrameMsg: *frame}
		err := json.Unmarshal(frame.RawParams, &pollerRegisterReq.Params)
		if err != nil {
			logUnmarshalError(frame, c)
			return err
		}

		s.AgentTracker.ProcessPollerRegister(*pollerRegisterReq)

	case protocol.MethodCheckMetricsPost:
		metricsPostReq := &protocol.MetricsPostRequest{FrameMsg: *frame}
		err := json.Unmarshal(frame.RawParams, &metricsPostReq.Params)
		if err != nil {
			logUnmarshalError(frame, c)
			return err
		}

		s.AgentTracker.ProcessCheckMetricsPost(*metricsPostReq)

	case protocol.MethodHeartbeatPost:
		log.WithField("remoteAddr", c.RemoteAddr()).Debug("It's alive")

	default:
		return types.Error{Msg: "Unsupported method: " + frame.Method}
	}

	return nil
}

func logUnmarshalError(frame *protocol.FrameMsg, c *utils.SmartConn) {
	log.WithFields(log.Fields{
		"rawParams":  frame.RawParams,
		"method":     frame.Method,
		"remoteAddr": c.RemoteAddr(),
	}).Warn("Failed to unmarshal raw params")
}

func watchForAgentErrors(ctx context.Context, agentErrors <-chan error, c *utils.SmartConn) {
	select {
	case err := <-agentErrors:
		log.WithField("remoteAddr", c.RemoteAddr()).Warn("Agent registration problem. Closing channel.", err)
		c.Close()

	case <-ctx.Done():
		return
	}

}
