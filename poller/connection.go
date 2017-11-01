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

	log "github.com/sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/rackerlabs/go-connect-tunnel"
)

var (
	metricsConnectionsAttempt = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "poller",
			Subsystem: "connection",
			Name:      "attempts",
			Help:      "Conveys the number of connection attempts per remote address",
		},
		[]string{
			metricLabelAddress,
		},
	)
	metricsConnectionsConnected = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "poller",
			Subsystem: "connection",
			Name:      "connected",
			Help:      "Conveys the number of successful connections per remote address",
		},
		[]string{
			metricLabelAddress,
		},
	)
	metricsConnectionsDisconnected = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "poller",
			Subsystem: "connection",
			Name:      "disconnected",
			Help:      "Conveys the number of disconnects per remote address",
		},
		[]string{
			metricLabelAddress,
		},
	)
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

	authenticated chan struct{}
}

func init() {
	metricsRegistry.MustRegister(metricsConnectionsAttempt)
	metricsRegistry.MustRegister(metricsConnectionsConnected)
	metricsRegistry.MustRegister(metricsConnectionsDisconnected)
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
		authenticated:     make(chan struct{}, 1),
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

func (conn *EleConnection) HasLatencyMeasurements() bool {
	return conn.session.HasLatencyMeasurements()
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

func (conn *EleConnection) String() string {
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

	counter, err := metricsConnectionsAttempt.GetMetricWithLabelValues(conn.address)
	if err == nil {
		counter.Inc()
	} else {
		log.WithField("err", err).Debug("Failed to get metricsConnectionsAttempt counter")
	}

	if err = conn.dial(config, tlsConfig); err != nil {
		return err
	}

	conn.session = NewSession(ctx, conn, conn.checksReconciler, config)
	log.WithFields(log.Fields{
		"prefix":         conn.GetLogPrefix(),
		"remote_address": conn.address,
	}).Info("Connected")

	counter, err = metricsConnectionsConnected.GetMetricWithLabelValues(conn.address)
	if err == nil {
		counter.Inc()
	} else {
		log.WithField("err", err).Debug("Failed to get metricsConnectionsConnected counter")
	}

	return nil
}

func (conn *EleConnection) dial(config *config.Config, tlsConfig *tls.Config) (err error) {
	nd := net.Dialer{Timeout: conn.connectionTimeout}

	var proxyConn net.Conn
	if config.ProxyUrl != nil {
		log.WithFields(log.Fields{
			"prefix": conn.GetLogPrefix(),
			"proxy":  config.ProxyUrl,
		}).Info("Connecting to agent/poller endpoint via proxy")
		proxyConn, err = tunnel.DialViaProxy(config.ProxyUrl, conn.address)
		if err != nil {
			return err
		}
	}

	if tlsConfig != nil {
		if proxyConn != nil {
			conn.conn = tls.Client(proxyConn, tlsConfig)
		} else {
			conn.conn, err = tls.DialWithDialer(&nd, "tcp", conn.address, tlsConfig)
			if err != nil {
				return err
			}
		}
	} else {
		if proxyConn != nil {
			conn.conn = proxyConn;
		} else {
			conn.conn, err = nd.Dial("tcp", conn.address)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Close closes the session
func (conn *EleConnection) Close() {
	if conn.conn != nil {
		log.WithFields(log.Fields{
			"prefix":         conn.GetLogPrefix(),
			"remote_address": conn.address,
		}).Info("Disconnected")

		counter, err := metricsConnectionsDisconnected.GetMetricWithLabelValues(conn.address)
		if err == nil {
			counter.Inc()
		} else {
			log.WithField("err", err).Debug("Failed to get metricsConnectionsDisconnected counter")
		}

		conn.conn.Close()
		conn.conn = nil
	}
	conn.GetSession().Close()
}

// Wait returns a channel that is populated when the connection is finished or closed.
func (conn *EleConnection) Done() <-chan struct{} {
	return conn.GetSession().Done()
}

func (conn *EleConnection) SetAuthenticated() {
	close(conn.authenticated)
}

func (conn *EleConnection) Authenticated() <-chan struct{} {
	return conn.authenticated
}
