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
	"crypto/tls"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"go/types"
	"golang.org/x/net/context"
	"net"
	"strconv"
)

type BasicServer struct {
	Certificate *tls.Certificate
	BindAddr    string

	AgentTracker
}

func (s *BasicServer) ApplyConfig(cfg config.EndpointConfig) error {

	cert, err := LoadCertificateFromConfig(cfg)
	if err != nil {
		return err
	}
	s.Certificate = cert

	bindAddr := cfg.BindAddr
	if bindAddr == "" {
		bindAddr = ":" + strconv.Itoa(config.DefaultPort)
	}
	s.BindAddr = bindAddr

	//TODO scan agents config dir

	return nil
}

func (s *BasicServer) ListenAndServe() error {
	tlsConfig := tls.Config{
		Certificates: []tls.Certificate{*s.Certificate},
	}

	listener, err := tls.Listen("tcp", s.BindAddr, &tlsConfig)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Infoln("Endpoint is accepting connections from pollers", listener.Addr())

	rootContext, _ := context.WithCancel(context.Background())

	for {
		conn, err := listener.Accept()
		log.Infoln("accepted...")
		if err != nil {
			log.Errorln("Failed to accept connection", err.Error())
		}
		go s.handleConnection(rootContext, conn)
	}
}

func (s *BasicServer) handleConnection(ctx context.Context, c net.Conn) {
	defer c.Close()
	log.Infoln("Handling connection", c.RemoteAddr())

	decoder := json.NewDecoder(c)
	encoder := json.NewEncoder(c)
	frame := &protocol.FrameMsg{}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case err := decoder.Decode(frame):
			if err != nil {
				log.Warnln("Failed to decode frame", c.RemoteAddr())
				return
			}

			log.Debugln("Received frame", frame)

			err = s.consumeFrame(ctx, c, frame, encoder)
			if err != nil {
				log.Warnln("Failed to consume frame", err)
				// assume the worst, get out, and close the connection
				return
			}
		}

	}
}

func (s *BasicServer) consumeFrame(ctx context.Context, c net.Conn, frame *protocol.FrameMsg, encoder *json.Encoder) error {
	switch frame.Method {
	case protocol.MethodHandshakeHello:
		handshakeReq := &protocol.HandshakeRequest{FrameMsg: *frame}
		json.Unmarshal(*frame.RawParams, &handshakeReq.Params)
		log.Infoln("RX handshake hello", handshakeReq)

		agentErrors := GetAgentTracker().ProcessHello(handshakeReq)
		go watchForAgentErrors(ctx, agentErrors, c)

		resp := &protocol.HandshakeResponse{}
		resp.Method = protocol.MethodEmpty
		resp.Id = frame.Id

		encoder.Encode(resp)
	default:
		return types.Error{Msg: "Unsupported method: " + frame.Method}
	}

	return nil
}

func watchForAgentErrors(ctx context.Context, agentErrors chan error, c net.Conn) {
	select {
	case err := <-agentErrors:
		log.WithField("remoteAddr", c.RemoteAddr()).Warn("Agent registration problem. Closing channel.", err)
		c.Close()

	case <-ctx.Done():
		return
	}

}
