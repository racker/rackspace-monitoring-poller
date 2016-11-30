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
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/config"
	"net"
	"sync"
	"time"
)

type ConnectionStream struct {
	ctx context.Context

	stopCh chan struct{}
	config *config.Config

	connsMu sync.Mutex
	conns   map[string]*Connection
	wg      sync.WaitGroup

	scheduler *Scheduler
}

var (
	ReconnectTimeout = 25 * time.Second
)

func NewConnectionStream(config *config.Config) *ConnectionStream {
	stream := &ConnectionStream{config: config}
	stream.ctx = context.Background()
	stream.conns = make(map[string]*Connection)
	stream.stopCh = make(chan struct{}, 1)
	stream.scheduler = NewScheduler("pzA", stream)
	go stream.scheduler.run()
	return stream
}

func (cs *ConnectionStream) GetConfig() *config.Config {
	return cs.config
}

func (cs *ConnectionStream) RegisterConnection(qry string, conn *Connection) {
	cs.connsMu.Lock()
	defer cs.connsMu.Unlock()
	cs.conns[qry] = conn
}

func (cs *ConnectionStream) Stop() {
	for _, conn := range cs.conns {
		conn.Close()
	}
	cs.stopCh <- struct{}{}
}

func (cs *ConnectionStream) StopNotify() chan struct{} {
	return cs.stopCh
}

func (cs *ConnectionStream) GetScheduler() *Scheduler {
	return cs.scheduler
}

func (cs *ConnectionStream) SendMetrics(crs *check.CheckResultSet) {
	for _, conn := range cs.conns {
		// TODO make this better
		conn.session.Send(check.NewMetricsPostRequest(crs))
		break
	}
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
	defer cs.wg.Done()
	for {
		conn := NewConnection(addr, cs.GetConfig().Guid, cs)
		err := conn.Connect(cs.ctx)
		if err != nil {
			goto error
		}
		cs.RegisterConnection(addr, conn)
		conn.Wait()
		goto new_connection
	error:
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
