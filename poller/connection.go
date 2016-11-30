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
	log "github.com/Sirupsen/logrus"
	"io"
	"net"
	"time"
)

type Connection struct {
	stream *ConnectionStream

	session *Session
	conn    io.ReadWriteCloser

	address string
	guid    string

	connectionTimeout time.Duration
}

func NewConnection(address string, guid string, stream *ConnectionStream) *Connection {
	return &Connection{
		address:           address,
		guid:              guid,
		stream:            stream,
		connectionTimeout: time.Duration(10) * time.Second,
	}
}

func (conn *Connection) GetStream() *ConnectionStream {
	return conn.stream
}

func (conn *Connection) SetReadDeadline(deadline time.Time) {
	socket, ok := conn.conn.(*net.TCPConn)
	if ok {
		socket.SetReadDeadline(deadline)
	}
}

func (conn *Connection) SetWriteDeadline(deadline time.Time) {
	socket, ok := conn.conn.(*net.TCPConn)
	if ok {
		socket.SetWriteDeadline(deadline)
	}
}

func (conn *Connection) Connect(ctx context.Context) error {
	log.Infof("Connecting to %s with timeout %v", conn.address, conn.connectionTimeout)
	nd := net.Dialer{Timeout: conn.connectionTimeout}
	conf := &tls.Config{InsecureSkipVerify: true}
	tlsConn, err := tls.DialWithDialer(&nd, "tcp", conn.address, conf)
	if err != nil {
		return err
	}
	log.Info("  ... Connected")
	conn.conn = tlsConn
	conn.session = newSession(ctx, conn)
	return nil
}

func (conn *Connection) Close() {
	conn.session.Close()
}

func (conn *Connection) Wait() {
	conn.session.Wait()
}
