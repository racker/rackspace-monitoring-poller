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
	"github.com/racker/rackspace-monitoring-poller/utils"
)

var (
	// ErrInvalidConnectionStream used when conneciton stream is not properly initialized
	ErrInvalidConnectionStream = errors.New("ConnectionStream has not been properly set up.  Re-initialize")
	// ErrUndefinedContext used when passed in context in Connect is undefined
	ErrUndefinedContext = errors.New("Context is undefined")
	// ErrCheckEmpty used when a check is nil or empty
	ErrCheckEmpty = errors.New("Check is empty")
	// CheckSpreadInMilliseconds sets up jitter time so as not
	// to send all requests at the same time
	CheckSpreadInMilliseconds = 30000
)

const (
	gracefulShutdownTimeout = 5 * time.Second
)

type LogPrefixGetter interface {
	GetLogPrefix() string
}

// ConnectionStream interface wraps the necessary information to
// register, connect, and send data in connections.
// It is the main factory for connection handling
type ConnectionStream interface {
	ChecksReconciler
	utils.EventSource

	SendMetrics(crs *check.ResultSet)
	Connect()
	Done() <-chan struct{}
}

// Connection interface wraps the methods required to manage a
// single connection.
type Connection interface {
	ConnectionHealthProvider
	LogPrefixGetter

	GetSession() Session
	SetReadDeadline(deadline time.Time)
	SetWriteDeadline(deadline time.Time)
	Connect(ctx context.Context, config *config.Config, tlsConfig *tls.Config) error
	Close()
	// Done returns a channel that is closed when the connection is finished or closed.
	Done() <-chan struct{}
	GetFarendWriter() io.Writer
	GetFarendReader() io.Reader
	GetGUID() string

	// SetAuthenticated should be invoked when a handshake response is successfully received
	SetAuthenticated()
	// Authenticated returns a channel that is closed when authenticated
	Authenticated() <-chan struct{}
}

type ConnectionHealthProvider interface {
	GetClockOffset() int64
	// HasLatencyMeasurements indicates if the value returned by GetLatency is ready to be used
	HasLatencyMeasurements() bool
	GetLatency() int64
}

// Session interface wraps the methods required to manage a
// session in a connection.  It includes authentication, request/
// response timeout management, and transferring data
type Session interface {
	ConnectionHealthProvider
	utils.EventSource

	Auth()
	Send(msg protocol.Frame)
	Respond(msg protocol.Frame)
	Close()
	Done() <-chan struct{}
}

// CheckScheduler arranges the periodic invocation of the given Check
type CheckScheduler interface {
	Schedule(ch check.Check)
	CancelCheck(ch check.Check)
}

// CheckExecutor facilitates running a check and consuming the CheckResultSet
type CheckExecutor interface {
	Execute(ch check.Check)
}

// ChecksReconciler is implemented by receivers that can either reconcile prepared checks during a commit or
// pre-validate the checks prior to committing.
type ChecksReconciler interface {
	// ReconcileChecks acts upon the given ChecksPrepared during a commit-phase.
	// The bulk of the processing is likely handled in an alternate go routine, so errors in the given
	// ChecksPrepared are handled but not reportable back to this caller. Use ValidateChecks prior to calling
	// this to pre-compute those errors.
	ReconcileChecks(cp ChecksPrepared)

	// Validate goes through the motions of ReconcileChecks in order to pre-validate consistency.
	// Unlike ReconcileChecks, this function should only require the manifest level of detail in the ActionableCheck
	// instances.
	// Returns an error upon finding the first entry that is not valid.
	ValidateChecks(cp ChecksPreparing) error
}

// Scheduler interface wraps the methods that schedule
// metric setup and sending
type Scheduler interface {
	ChecksReconciler

	// Reset will stop and de-schedule all checks, but leave this scheduler available for reconciling new checks.
	Reset()
	// Close is only used for isolated, such as unit test, use of the Scheduler in order to close out go routines
	Close()

	SendMetrics(crs *check.ResultSet)
	GetZoneID() string
	GetContext() (ctx context.Context, cancel context.CancelFunc)
	GetScheduledChecks() []check.Check
}

type ConnectionFactory func(address string, guid string, checksReconciler ChecksReconciler) Connection

type ConnectionsByHost map[string]Connection

type FrameMsgError struct {
	Frame *protocol.FrameMsg
	Error error
}
