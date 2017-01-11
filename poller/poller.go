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

// Package poller contains the poller/agent side connectivity and coordination logic.
package poller

import (
	"context"
	"crypto/tls"
	"io"
	"time"

	"errors"

	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/protocol"
)

var (
	// ReconnectTimeout sets up 25 second timeout for reconnection
	ReconnectTimeout = 25 * time.Second
	// ErrInvalidConnectionStream used when conneciton stream is not properly initialized
	ErrInvalidConnectionStream = errors.New("ConnectionStream has not been properly set up.  Re-initialize")
	// ErrNoConnections used when no connections were set up in the stream
	ErrNoConnections = errors.New("No connections")
	// ErrUndefinedContext used when passed in context in Connect is undefined
	ErrUndefinedContext = errors.New("Context is undefined")
	// ErrCheckEmpty used when a check is nil or empty
	ErrCheckEmpty = errors.New("Check is empty")
	// CheckSpreadInMilliseconds sets up jitter time so as not
	// to send all requests at the same time
	CheckSpreadInMilliseconds = 30000
)

// ConnectionStream interface wraps the necessary information to
// register, connect, and send data in connections.
// It is the main factory for connection handling
type ConnectionStream interface {
	GetConfig() *config.Config
	RegisterConnection(qry string, conn Connection) error
	Stop()
	StopNotify() chan struct{}
	GetSchedulers() map[string]Scheduler
	GetContext() context.Context
	SendMetrics(crs *check.ResultSet) error
	Connect()
	WaitCh() <-chan struct{}
	GetConnections() map[string]Connection
}

// Connection interface wraps the methods required to manage a
// single connection.
type Connection interface {
	GetStream() ConnectionStream
	GetSession() Session
	SetReadDeadline(deadline time.Time)
	SetWriteDeadline(deadline time.Time)
	Connect(ctx context.Context, tlsConfig *tls.Config) error
	Close()
	Wait()
	GetConnection() io.ReadWriteCloser
	GetGUID() string
}

// Session interface wraps the methods required to manage a
// session in a connection.  It includes authentication, request/
// response timeout management, and transferring data
type Session interface {
	Auth()
	Send(msg protocol.Frame)
	Respond(msg protocol.Frame)
	SetHeartbeatInterval(timeout uint64)
	GetReadDeadline() time.Time
	GetWriteDeadline() time.Time
	Close()
	Wait()
}

// Scheduler interface wraps the methods that schedule
// metric setup and sending
type Scheduler interface {
	GetInput() chan protocol.Frame
	Close()
	SendMetrics(crs *check.ResultSet)
	Register(ch check.Check) error
	RunFrameConsumer()
	GetZoneID() string
	GetContext() (ctx context.Context, cancel context.CancelFunc)
	GetChecks() map[string]check.Check
}
