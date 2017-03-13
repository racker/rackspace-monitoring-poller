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
	"github.com/racker/rackspace-monitoring-poller/config"
)

// EleConnection implements Connection
// See Connection interface for more information
type EleConnection struct {
	checksReconciler ChecksReconciler

	session Session
	conn    io.ReadWriteCloser

	address string
	guid    string

	connectionTimeout time.Duration
}

// NewConnection instantiates a new EleConnection
// It sets up the address, unique guid, and connection timeout
// for this conneciton
func NewConnection(address string, guid string, checksReconciler ChecksReconciler) Connection {
	return &EleConnection{
		address:           address,
		guid:              guid,
		checksReconciler:  checksReconciler,
		connectionTimeout: time.Duration(10) * time.Second,
	}
}

// GetGUID retrieves connection's guid
func (conn *EleConnection) GetGUID() string {
	return conn.guid
}

// GetFarendWriter gets a writer directed towards the endpoint server
func (conn *EleConnection) GetFarendWriter() io.Writer {
	return conn.conn
}

// GetFarendReader gets a reader to consume from the endpoint server
func (conn *EleConnection) GetFarendReader() io.Reader {
	return conn.conn
}

// GetSession retrieves connection's session
func (conn *EleConnection) GetSession() Session {
	return conn.session
}

func (conn *EleConnection) GetClockOffset() int64 {
	return conn.session.GetClockOffset()
}

func (conn *EleConnection) GetLatency() int64 {
	return conn.session.GetLatency()
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

func (conn *EleConnection) GetLogPrefix() string {
	return conn.address
}

// Connect sets up a tcp connection with connection defined address
// and passed in tlsConfig
// If context is not set, ErrUndefinedContext is returned
// The end result of this function is a usable connection ready to
// send data.
func (conn *EleConnection) Connect(ctx context.Context, config *config.Config, tlsConfig *tls.Config) error {
	if ctx == nil {
		return ErrUndefinedContext
	}
	log.WithFields(log.Fields{
		"prefix":  conn.GetLogPrefix(),
		"timeout": conn.connectionTimeout,
	}).Info("Connecting to agent/poller endpoint")
	nd := net.Dialer{Timeout: conn.connectionTimeout}
	tlsConn, err := tls.DialWithDialer(&nd, "tcp", conn.address, tlsConfig)
	if err != nil {
		return err
	}
	conn.conn = tlsConn
	conn.session = NewSession(ctx, conn, conn.checksReconciler, config)
	log.WithFields(log.Fields{
		"prefix":         conn.GetLogPrefix(),
		"remote_address": tlsConn.RemoteAddr(),
	}).Info("Connected")
	return nil
}

// Close closes the session
func (conn *EleConnection) Close() {
	if conn.conn != nil {
		conn.conn.Close()
	}
	conn.GetSession().Close()
}

// Wait returns a channel that is populated when the connection is finished or closed.
func (conn *EleConnection) Wait() <-chan struct{} {
	return conn.GetSession().Wait()
}
