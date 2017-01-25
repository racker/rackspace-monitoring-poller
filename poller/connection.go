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
	"crypto/tls"
	"io"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
)

// EleConnection implements Connection
// See Connection interface for more information
type EleConnection struct {
	stream ConnectionStream

	session Session
	conn    io.ReadWriteCloser

	address string
	guid    string

	connectionTimeout time.Duration
}

// NewConnection instantiates a new EleConnection
// It sets up the address, unique guid, and connection timeout
// for this conneciton
func NewConnection(address string, guid string, stream ConnectionStream) Connection {
	return &EleConnection{
		address:           address,
		guid:              guid,
		stream:            stream,
		connectionTimeout: time.Duration(10) * time.Second,
	}
}

// GetGUID retrieves connection's guid
func (conn *EleConnection) GetGUID() string {
	return conn.guid
}

// GetConnection returns ReadWriteCloser used for streaming data
func (conn *EleConnection) GetConnection() io.ReadWriteCloser {
	return conn.conn
}

// GetStream retrieves connection's stream
func (conn *EleConnection) GetStream() ConnectionStream {
	return conn.stream
}

// GetSession retrieves connection's session
func (conn *EleConnection) GetSession() Session {
	return conn.session
}

// SetReadDeadline sets up the read deadline for a socket.
// read deadline - time spent reading response for future calls
func (conn *EleConnection) SetReadDeadline(deadline time.Time) {
	socket, ok := conn.conn.(*net.TCPConn)
	if ok {
		socket.SetReadDeadline(deadline)
	}
}

// SetWriteDeadline sets up the write deadline for a socket.
// write deadline - time spent writing request to socket for future calls
func (conn *EleConnection) SetWriteDeadline(deadline time.Time) {
	socket, ok := conn.conn.(*net.TCPConn)
	if ok {
		socket.SetWriteDeadline(deadline)
	}
}

// Connect sets up a tcp connection with connection defined address
// and passed in tlsConfig
// If context is not set, ErrUndefinedContext is returned
// The end result of this function is a usable connection ready to
// send data.
func (conn *EleConnection) Connect(ctx context.Context, tlsConfig *tls.Config) error {
	if ctx == nil {
		return ErrUndefinedContext
	}
	log.WithFields(log.Fields{
		"address": conn.address,
		"timeout": conn.connectionTimeout,
	}).Info("Connecting to agent/poller endpoint")
	nd := net.Dialer{Timeout: conn.connectionTimeout}
	tlsConn, err := tls.DialWithDialer(&nd, "tcp", conn.address, tlsConfig)
	if err != nil {
		return err
	}
	log.Info("  ... Connected")
	conn.conn = tlsConn
	conn.session = newSession(ctx, conn)
	return nil
}

// Close closes the session
func (conn *EleConnection) Close() {
	if conn.conn != nil {
		conn.conn.Close()
	}
	conn.GetSession().Close()
}

// Wait sets the connection session to wait for a new request
func (conn *EleConnection) Wait() {
	conn.GetSession().Wait()
}
