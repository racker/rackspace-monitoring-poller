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
package types

import (
	"context"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/check"
	"net"
	"sync"
	"time"
)

type ConnectionStream struct {
	config *Config

	connsMu sync.Mutex
	conns   map[string]*Connection
	wg      sync.WaitGroup

	scheduler *Scheduler
}

func NewConnectionStream(config *Config) *ConnectionStream {
	stream := &ConnectionStream{config: config}
	stream.conns = make(map[string]*Connection)
	stream.scheduler = NewScheduler("pzA", stream)
	go stream.scheduler.run()
	return stream
}

func (cs *ConnectionStream) GetConfig() *Config {
	return cs.config
}

func (cs *ConnectionStream) RegisterConnection(qry string, conn *Connection) {
	cs.connsMu.Lock()
	defer cs.connsMu.Unlock()
	cs.conns[qry] = conn
}

func (cs *ConnectionStream) GetScheduler() *Scheduler {
	return cs.scheduler
}

func (cs *ConnectionStream) SendMetrics(crs *check.CheckResultSet) {
	for _, conn := range cs.conns {
		// TODO make this better
		conn.session.Send(NewMetricsPostRequest(crs))
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

func (cs *ConnectionStream) Wait() {
	cs.wg.Wait()
}

func (cs *ConnectionStream) connectBySrv(qry string) {
	_, addrs, err := net.LookupSRV("", "", qry)
	if err != nil {
		log.Errorf("SRV Lookup Failure", err)
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
	reconnectTimeout := time.Duration(25 * time.Second)
	for {
		conn := NewConnection(addr, cs.GetConfig().Guid, cs)
		err := conn.Connect(context.Background())
		if err != nil {
			goto error
		}
		cs.RegisterConnection(addr, conn)
		conn.Wait()
		goto new_connection
	error:
		log.Errorf("Error: %v", err)
	new_connection:
		log.Infof("  connection sleeping %v", reconnectTimeout)
		time.Sleep(reconnectTimeout)
	}
}
