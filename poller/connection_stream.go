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
	"crypto/x509"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/jpillora/backoff"
	log "github.com/sirupsen/logrus"

	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/utils"
)

const (
	metricsChannelSize       = 100
	registrationsChannelSize = 1
	// EventTypeRegister has a target of Connection
	EventTypeRegister = "register"
	// EventTypeDeregister has a target of Connection
	EventTypeDeregister = "deregister"
	// EventTypeDroppedMetric has a target of *check.ResultSet
	EventTypeDroppedMetric = "dropped"
	// EventTypeAllConnectionsLost indicates that all endpoint connections were lost
	EventTypeAllConnectionsLost = "allConnectionsLost"
)

// EleConnectionStream implements ConnectionStream
// See ConnectionStream for more information
type EleConnectionStream struct {
	LogPrefixGetter
	utils.EventConsumerRegistry

	ctx     context.Context
	cancel  context.CancelFunc
	rootCAs *x509.CertPool

	config *config.Config

	connectionFactory ConnectionFactory
	conns             ConnectionsByHost
	wg                sync.WaitGroup

	registrations chan *connectionRegistration
	metricsToSend chan *check.ResultSet

	// map is the private zone ID as a string
	schedulers map[string]Scheduler

	metricsDistributor *MetricsDistributor
}

type connectionRegistration struct {
	register bool
	qry      string
	conn     Connection
}

// NewConnectionStream instantiates a new EleConnectionStream
// It sets up the contexts and the starts the schedulers based on configured private zones
func NewConnectionStream(ctx context.Context, config *config.Config, rootCAs *x509.CertPool) ConnectionStream {
	return NewCustomConnectionStream(ctx, config, rootCAs, nil)
}

// NewCustomConnectionStream is a variant of NewConnectionStream that allows providing a customized ConnectionFactory
func NewCustomConnectionStream(ctx context.Context, config *config.Config, rootCAs *x509.CertPool, connectionFactory ConnectionFactory) ConnectionStream {
	if connectionFactory == nil {
		connectionFactory = NewConnection
	}
	stream := &EleConnectionStream{
		config:             config,
		rootCAs:            rootCAs,
		schedulers:         make(map[string]Scheduler),
		connectionFactory:  connectionFactory,
		metricsToSend:      make(chan *check.ResultSet, metricsChannelSize),
		registrations:      make(chan *connectionRegistration, registrationsChannelSize),
		metricsDistributor: NewMetricsDistributor(ctx, config),
	}
	stream.ctx, stream.cancel = context.WithCancel(ctx)
	stream.conns = make(ConnectionsByHost)
	for _, pz := range config.ZoneIds {
		stream.schedulers[pz] = NewScheduler(pz, stream)
	}

	go stream.runRegistrationMetricsCoordinator()

	return stream
}

// GetLogPrefix returns the log prefix for this module
func (cs *EleConnectionStream) GetLogPrefix() string {
	return "stream"
}

// getRegisteredConnectionNames returns the registered connection names
func (cs *EleConnectionStream) getRegisteredConnectionNames() []string {
	names := []string{}
	for _, conn := range cs.conns {
		names = append(names, conn.GetLogPrefix())
	}
	return names
}

func (cs *EleConnectionStream) runRegistrationMetricsCoordinator() {
	log.WithField("prefix", cs.GetLogPrefix()).Debug("Registration/Metrics Coordinator starting")
	defer log.WithField("prefix", cs.GetLogPrefix()).Debug("Registration/Metrics Coordinator exiting")

	for {
		select {
		case <-cs.ctx.Done():
			return

		case crs := <-cs.metricsToSend:
			cs.sendMetrics(crs)

		case reg := <-cs.registrations:
			if reg.register {
				cs.registerConnection(reg.qry, reg.conn)
			} else {
				cs.deregisterConnection(reg.qry, reg.conn)
			}
		}
	}
}

func (cs *EleConnectionStream) registerConnection(qry string, conn Connection) {
	cs.conns[qry] = conn
	log.WithFields(log.Fields{
		"prefix":      cs.GetLogPrefix(),
		"connections": cs.getRegisteredConnectionNames(),
	}).Debug("After registering, currently registered connections")
	cs.EmitEvent(utils.NewEvent(EventTypeRegister, conn))
}

func (cs *EleConnectionStream) deregisterConnection(qry string, conn Connection) {
	if _, exists := cs.conns[qry]; exists {
		delete(cs.conns, qry)
		log.WithFields(log.Fields{
			"prefix":      cs.GetLogPrefix(),
			"connections": cs.conns,
		}).Debug("After deregistering, currently registered connections")

		cs.EmitEvent(utils.NewEvent(EventTypeDeregister, conn))
		if len(cs.conns) == 0 {
			log.WithFields(log.Fields{
				"prefix": cs.GetLogPrefix(),
			}).Info("Lost all connections to endpoints")

			for _, scheduler := range cs.schedulers {
				scheduler.Reset()
			}

			cs.EmitEvent(utils.NewEvent(EventTypeAllConnectionsLost, nil))

			newGuid := generatePollerGuid()
			log.WithFields(log.Fields{
				"prefix":   cs.GetLogPrefix(),
				"previous": cs.config.Guid,
				"new":      newGuid,
			}).Info("Re-generating GUID to start new logical poller session")
			cs.config.Guid = newGuid
		}
	} else {
		log.WithFields(log.Fields{
			"prefix":     cs.GetLogPrefix(),
			"connection": conn,
		}).Debug("Saw deregistration, but connection wasn't fully ready.")
	}
}

// ReconcileChecks routes the ChecksPreparation to its schedulers.
func (cs *EleConnectionStream) ReconcileChecks(cp ChecksPrepared) {
	for _, sched := range cs.schedulers {
		sched.ReconcileChecks(cp)
	}
}

func (cs *EleConnectionStream) ValidateChecks(cp ChecksPreparing) error {
	for _, sched := range cs.schedulers {
		err := sched.ValidateChecks(cp)
		if err != nil {
			log.WithFields(log.Fields{
				"prefix":    cs.GetLogPrefix(),
				"scheduler": sched,
				"cp":        cp,
				"err":       err,
			}).Warn("Scheduler was not able to validate check preparation")
			return err
		}
	}

	return nil
}

// SendMetrics sends a CheckResultSet via the first connection it can
// retrieve in the connection list
func (cs *EleConnectionStream) SendMetrics(crs *check.ResultSet) {
	cs.metricsToSend <- crs

	cs.metricsDistributor.Distribute(crs)
}

func (cs *EleConnectionStream) sendMetrics(crs *check.ResultSet) {
	if cs.conns == nil || len(cs.conns) == 0 {
		crsJson, _ := json.Marshal(crs)
		log.WithFields(log.Fields{
			"prefix":    cs.GetLogPrefix(),
			"resultSet": string(crsJson),
		}).Warn("No connections are available for sending metrics")

		cs.EmitEvent(utils.NewEvent(EventTypeDroppedMetric, crs))

		return
	}

	if conn := cs.conns.ChooseBest(); conn != nil {
		conn.GetSession().Send(check.NewMetricsPostRequest(crs, conn.GetClockOffset()))
	}

}

// Connect connects to configured endpoints.
// There are 2 ways to connect:
// 1. You can utilize SRV records defined in the configuration
// to dynamically find endpoints
// 2. You can explicitly specify endpoint addresses and connect
// to them directly
// DEFAULT: Using SRV records
func (cs *EleConnectionStream) Connect() {
	if cs.config.UseSrv {
		for _, qry := range cs.config.SrvQueries {
			cs.wg.Add(1)
			go cs.connectBySrv(qry)
		}
	} else {
		for _, addr := range cs.config.Addresses {
			cs.wg.Add(1)
			go cs.runHostConnection(addr)
		}
	}

	cs.metricsDistributor.Start()
}

// Done provides a channel for waiting on connection closure
func (cs *EleConnectionStream) Done() <-chan struct{} {
	c := make(chan struct{}, 1)
	go func() {
		cs.wg.Wait()
		close(c)
	}()
	return c
}

func (cs *EleConnectionStream) connectBySrv(qry string) {
	log.WithFields(log.Fields{
		"prefix": cs.GetLogPrefix(),
		"query":  qry,
	}).Debug("Resolving by service")
	_, addrs, err := net.LookupSRV("", "", qry)
	if err != nil {
		log.WithFields(log.Fields{
			"prefix": cs.GetLogPrefix(),
		}).Errorf("SRV Lookup Failure : %v", err)
		cs.wg.Done()
		return
	}
	if len(addrs) == 0 {
		log.WithFields(log.Fields{
			"prefix": cs.GetLogPrefix(),
		}).Error("no addresses returned")
		cs.wg.Done()
		return
	}
	addr := net.JoinHostPort(addrs[0].Target, fmt.Sprintf("%v", addrs[0].Port))
	log.WithFields(log.Fields{
		"prefix": cs.GetLogPrefix(),
		"query":  qry,
		"addr":   addr,
	}).Debug("Resolved service")
	cs.runHostConnection(addr)
}

func (cs *EleConnectionStream) runHostConnection(addr string) {

	log.WithFields(log.Fields{
		"prefix": cs.GetLogPrefix(),
		"addr":   addr,
	}).Debug("Connecting by address")
	defer log.WithFields(log.Fields{
		"prefix": cs.GetLogPrefix(),
		"addr":   addr,
	}).Debug("Connection exiting")
	defer cs.wg.Done()

	b := &backoff.Backoff{
		Min:    cs.config.ReconnectMinBackoff,
		Max:    cs.config.ReconnectMaxBackoff,
		Factor: cs.config.ReconnectFactorBackoff,
		Jitter: true,
	}

	var maxConnectionAgeChan <-chan time.Time

reconnect:
	for {
		conn := cs.connectionFactory(addr, cs.config.Guid, cs)
		if conn == nil {
			log.WithFields(log.Fields{
				"prefix": cs.GetLogPrefix(),
			}).Error("Connection factory provided nil")
			return
		}
		pendingAuth := conn.Authenticated()
		err := conn.Connect(cs.ctx, cs.config, cs.buildTLSConfig(addr))
		if err != nil {
			goto conn_error
		}

		// Successful connection. reset backoff
		b.Reset()

		if cs.config.ProxyUrl != nil && cs.config.MaxProxyConnectionAge > 0 {
			maxConnectionAge := cs.config.MaxProxyConnectionAge +
				time.Duration(float64(cs.config.MaxProxyConnectionAgeJitter)*rand.Float64())
			maxConnectionAgeChan = time.After(maxConnectionAge)

			log.WithField("maxConnectionAge", maxConnectionAge).Debug("Limiting max connection age for proxied connection")
		}

		for {
			select {
			case <-pendingAuth:
				cs.registrations <- &connectionRegistration{
					register: true,
					qry:      addr,
					conn:     conn,
				}
				// no need to proceed on this closed channel
				pendingAuth = nil

			case <-cs.ctx.Done():
				// external cancellation
				conn.Close()
				return

			case <-maxConnectionAgeChan:
				conn.Close()

			case <-conn.Done():
				// connection closed
				cs.registrations <- &connectionRegistration{
					register: false,
					qry:      addr,
					conn:     conn,
				}
				goto new_connection
			}
		}

	conn_error:
		log.WithFields(log.Fields{
			"prefix":  cs.GetLogPrefix(),
			"address": addr,
		}).Errorf("Error: %v", err)
	new_connection:
		sleepDuration := b.Duration()
		log.WithFields(log.Fields{
			"prefix":  cs.GetLogPrefix(),
			"address": addr,
			"timeout": sleepDuration,
		}).Debug("Connection sleeping")
		for {
			select {
			case <-cs.ctx.Done():
				log.WithFields(log.Fields{
					"prefix":  cs.GetLogPrefix(),
					"address": addr,
				}).Debug("Connection cancelled")
				return
			case <-time.After(sleepDuration):
				log.WithFields(log.Fields{
					"prefix":  cs.GetLogPrefix(),
					"address": addr,
				}).Debug("Reconnecting")
				continue reconnect
			}
		}
	}
}

func (cs *EleConnectionStream) buildTLSConfig(addr string) *tls.Config {
	host, _, _ := net.SplitHostPort(addr)

	if !config.IsUsingCleartext() {
		conf := &tls.Config{
			InsecureSkipVerify: cs.rootCAs == nil,
			ServerName:         host,
			RootCAs:            cs.rootCAs,
		}
		return conf
	} else {
		return nil
	}
}

// ChooseBest selects the best of its connections for posting metrics, etc.
// Returns nil if no connections were present.
func (conns ConnectionsByHost) ChooseBest() Connection {
	var minLatency int64 = math.MaxInt64
	var best Connection

	for _, conn := range conns {
		if conn.HasLatencyMeasurements() {
			latency := conn.GetLatency()
			if latency < minLatency {
				minLatency = latency
				best = conn
			}
		}
	}

	return best
}
