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
	"fmt"
	"net"
	"sync"
	"time"

	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/config"
)

type ConnectionStreamInterface interface {
	GetConfig() *config.Config
	RegisterConnection(qry string, conn ConnectionInterface) error
	Stop()
	StopNotify() chan struct{}
	GetScheduler() *Scheduler
	SendMetrics(crs *check.CheckResultSet) error
	Connect()
	WaitCh() <-chan struct{}
	GetConnections() map[string]ConnectionInterface
}

type ConnectionStream struct {
	ctx     context.Context
	rootCAs *x509.CertPool

	stopCh chan struct{}
	config *config.Config

	connsMu sync.Mutex
	conns   map[string]ConnectionInterface
	wg      sync.WaitGroup

	// map is the private zone ID as a string
	scheduler map[string]*Scheduler
}

var (
	ReconnectTimeout = 25 * time.Second
)

func NewConnectionStream(config *config.Config, rootCAs *x509.CertPool) ConnectionStreamInterface {
	stream := &ConnectionStream{
		config:    config,
		rootCAs:   rootCAs,
		scheduler: make(map[string]*Scheduler),
	}
	stream.ctx = context.Background()
	stream.conns = make(map[string]ConnectionInterface)
	stream.stopCh = make(chan struct{}, 1)
	for _, pz := range config.ZoneIds {
		stream.scheduler[pz] = NewScheduler(pz, stream)
		go stream.scheduler[pz].runFrameConsumer()
	}
	return stream
}

func (cs *ConnectionStream) GetConfig() *config.Config {
	return cs.config
}

func (cs *ConnectionStream) RegisterConnection(qry string, conn ConnectionInterface) error {
	cs.connsMu.Lock()
	defer cs.connsMu.Unlock()
	if cs.conns == nil {
		return errors.New("ConnectionStream has not been properly set up.  Re-initialize")
	}
	cs.conns[qry] = conn
	log.Warningf("%v", cs.conns)
	return nil
}

func (cs *ConnectionStream) GetConnections() map[string]ConnectionInterface {
	return cs.conns
}

func (cs *ConnectionStream) Stop() {
	if cs.conns != nil {
		for _, conn := range cs.conns {
			conn.Close()
		}
		cs.stopCh <- struct{}{}
	}
}

func (cs *ConnectionStream) StopNotify() chan struct{} {
	log.Info("stop notify")
	return cs.stopCh
}

func (cs *ConnectionStream) GetScheduler() map[string]*Scheduler {
	return cs.scheduler
}

func (cs *ConnectionStream) SendMetrics(crs *check.CheckResultSet) error {
	if cs.conns == nil {
		return errors.New("No connections")
	}
	for _, conn := range cs.conns {
		// TODO make this better
		conn.GetConnection().session.Send(check.NewMetricsPostRequest(crs))
		break
	}
	return nil

}

func (cs *ConnectionStream) Connect() {
	if cs.GetConfig().UseSrv {
		for _, qry := range cs.GetConfig().SrvQueries {
			cs.wg.Add(1)
			go cs.connectBySrv(qry)
		}
	} else {
		for _, addr := range cs.GetConfig().Addresses {
			cs.wg.Add(1)
			go cs.connectByHost(addr)
		}
	}
}

func (cs *ConnectionStream) WaitCh() <-chan struct{} {
	c := make(chan struct{}, 1)
	go func() {
		cs.wg.Wait()
		c <- struct{}{}
	}()
	return c
}

func (cs *ConnectionStream) connectBySrv(qry string) {
	_, addrs, err := net.LookupSRV("", "", qry)
	if err != nil {
		log.Errorf("SRV Lookup Failure : %v", err)
		return
	}
	if len(addrs) == 0 {
		log.Error("No addresses returned")
		return
	}
	addr := fmt.Sprintf("%s:%v", addrs[0].Target, addrs[0].Port)
	cs.connectByHost(addr)
}

func (cs *ConnectionStream) connectByHost(addr string) {
	var csi ConnectionStreamInterface = cs
	defer cs.wg.Done()
	for {
		conn := NewConnection(addr, csi.GetConfig().Guid, cs)
		err := conn.Connect(cs.ctx, cs.buildTlsConfig(addr))
		if err != nil {
			goto conn_error
		}
		err = cs.RegisterConnection(addr, conn)
		if err != nil {
			goto conn_error
		}
		conn.Wait()
		goto new_connection
	conn_error:
		log.Errorf("Error: %v", err)
	new_connection:
		log.Debugf("  connection sleeping %v", ReconnectTimeout)
		for {
			select {
			case <-cs.ctx.Done():
				log.Infof("connection close")
				return
			case <-time.After(ReconnectTimeout):
				break
			}
		}
	}
}

func (cs *ConnectionStream) buildTlsConfig(addr string) *tls.Config {
	host, _, _ := net.SplitHostPort(addr)
	conf := &tls.Config{
		InsecureSkipVerify: cs.rootCAs == nil,
		ServerName:         host,
		RootCAs:            cs.rootCAs,
	}
	return conf
}
