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
		go s.runConnectionHandler(rootContext, conn)
	}
}

func (s *BasicServer) runConnectionHandler(ctx context.Context, c net.Conn) {
	defer c.Close()
	log.WithField("remoteAddr", c.RemoteAddr()).Info("Started handling connection")
	defer log.WithField("remoteAddr", c.RemoteAddr()).Info("Stopped handling connection")

	smartC := utils.NewSmartConn(c)

	smartC.ReadKeepalive = ExpectedAgentHeartbeatSec * time.Second
	smartC.WriteAllowance = ConnectionWriteAllowance
	// this is also where we could setup endpoint->agent heartbeats
	err := smartC.Start()
	if err != nil {
		log.WithField("remoteAddr", c.RemoteAddr()).Errorln("Failed to setup read keepalives", err)
		return
	}

	var frames = make(chan *protocol.FrameMsg, 10)
	go s.runFrameDecoder(smartC, frames)

	for {
		select {
		case <-ctx.Done():
			log.WithField("remoteAddr", c.RemoteAddr()).Info("Done handling connection due to context being done", ctx.Err())
			return

		case frame := <-frames:
			if frame.IsFinished() {
				log.WithField("remoteAddr", c.RemoteAddr()).Debug("Handling finished frame")
				s.AgentTracker.CloseAgentByAddr(c.RemoteAddr())
				return
			}
			err := s.handleFrame(ctx, smartC, frame)
			if err != nil {
				log.Warnln("Failed to consume frame", err)
				// assume the worst, get out, and close the connection
				return
			}

		}

	}
}

// runFrameDecoder needs to be spun off since json.Decoder is a blocking read
func (s *BasicServer) runFrameDecoder(c *utils.SmartConn, frames chan<- *protocol.FrameMsg) {
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

		log.WithFields(log.Fields{
			"id":     frame.Id,
			"method": frame.Method,
			"from":   c.RemoteAddr(),
		}).Debug("RECV frame")
		frames <- &frame
	}
	frames <- protocol.NewFinishedFrame()
}

func (s *BasicServer) handleFrame(ctx context.Context, c *utils.SmartConn, frame protocol.Frame) error {
	log.WithFields(log.Fields{
		"remoteAddr": c.RemoteAddr(),
		"msgId":      frame.GetId(),
		"source":     frame.GetSource(),
		"method":     frame.GetMethod(),
	}).Debug("Consuming frame")

	switch frame.GetMethod() {
	case protocol.MethodHandshakeHello:
		params := &protocol.HandshakeParameters{}
		err := json.Unmarshal(frame.GetRawParams(), params)
		if err != nil {
			logUnmarshalError(c, frame)
			return err
		}

		sendHandshakeResponse(c, frame)

		agentErrors := s.AgentTracker.NewAgentFromHello(frame, params, c)
		go waitOnAgentError(ctx, agentErrors, c)

	case protocol.MethodCheckMetricsPost:
		params := &protocol.MetricsPostRequestParams{}

		err := json.Unmarshal(frame.GetRawParams(), params)
		if err != nil {
			logUnmarshalError(c, frame)
			return err
		}

		s.AgentTracker.handleCheckMetricsPost(&metricsPost{
			params: params,
			frame:  frame,
		})

	case protocol.MethodHeartbeatPost:
		log.WithField("remoteAddr", c.RemoteAddr()).Debug("RECV heartbeat")
		sendHeartbeatResponse(c, frame)

	case "": // response
		log.WithField("remoteAddr", c.RemoteAddr()).Debug("RECV response")
		s.AgentTracker.handleResponse(frame)

	default:
		return types.Error{Msg: "Unsupported method: " + frame.GetMethod()}
	}

	return nil
}

func sendHeartbeatResponse(c *utils.SmartConn, frame protocol.Frame) {
	resp := &protocol.HeartbeatResponse{}
	resp.Method = protocol.MethodEmpty
	resp.Id = frame.GetId()
	resp.Result.Timestamp = utils.NowTimestampMillis()

	log.WithField("resp", resp).Debug("SEND heartbeat resp")
	c.WriteJSON(resp)
}

func sendHandshakeResponse(c *utils.SmartConn, frame protocol.Frame) {
	resp := &protocol.HandshakeResponse{}
	resp.Method = protocol.MethodEmpty
	resp.Id = frame.GetId()
	resp.Result.HeartbeatInterval = ExpectedAgentHeartbeatSec * 1000

	log.WithField("resp", resp).Debug("SEND handshake resp")
	c.WriteJSON(resp)
}

func logUnmarshalError(c *utils.SmartConn, frame protocol.Frame) {
	log.WithFields(log.Fields{
		"rawParams":  frame.GetRawParams(),
		"method":     frame.GetMethod(),
		"remoteAddr": c.RemoteAddr(),
	}).Warn("Failed to unmarshal raw params")
}

func waitOnAgentError(ctx context.Context, agentErrors <-chan error, c *utils.SmartConn) {
	select {
	case err := <-agentErrors:
		log.WithFields(log.Fields{
			"remoteAddr": c.RemoteAddr(),
			"err":        err,
		}).Warn("Agent registration problem. Closing channel.")
		c.Close()

	case <-ctx.Done():
		return
	}
}
