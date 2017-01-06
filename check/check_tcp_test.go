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
package check_test

import (
	"context"
	"fmt"
	log "github.com/Sirupsen/logrus"
	check "github.com/racker/rackspace-monitoring-poller/check"
	"net"
	"sync"
	"testing"
	"time"
)

type BannerServer struct {
	HandleConnection func(conn net.Conn)

	waitGroup *sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewBannerServer() *BannerServer {
	server := &BannerServer{}
	server.waitGroup = &sync.WaitGroup{}
	server.ctx, server.cancel = context.WithCancel(context.Background())
	server.HandleConnection = server.serve
	server.waitGroup.Add(1)
	return server
}

func (s *BannerServer) Stop() {
	s.cancel()
	s.waitGroup.Wait()
}

func (s *BannerServer) Serve(listener *net.TCPListener) {
	defer s.waitGroup.Done()
	for {
		select {
		case <-s.ctx.Done():
			log.Debug("stopping listening on", listener.Addr())
			listener.Close()
			return
		default:
		}
		listener.SetDeadline(time.Now().Add(10 * time.Second))
		conn, err := listener.AcceptTCP()
		if nil != err {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
		}
		log.Debug(conn.RemoteAddr(), "connected")
		s.waitGroup.Add(1)
		go s.serve(conn)
	}
}

func (s *BannerServer) serve(conn net.Conn) {
	defer s.waitGroup.Done()
	defer conn.Close()
	for {
		select {
		case <-s.ctx.Done():
			log.Debug("disconnecting", conn.RemoteAddr())
			return
		default:
		}
		buf := make([]byte, 4096)
		conn.SetDeadline(time.Now().Add(1e9))
		conn.Write([]byte("SSH-2.0-OpenSSH_7.3\n"))
		conn.SetDeadline(time.Now().Add(1 * time.Second))
		if _, err := conn.Read(buf); nil != err {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			return
		}

	}
}

func ValidateMetrics(t *testing.T, metrics []string, cr *check.CheckResult) {
	for _, metricName := range metrics {
		if metric := cr.GetMetric(metricName); metric == nil {
			log.Fatal("metric " + metricName + " does not exist")
		}
	}

}

func TestTCPRunSuccess(t *testing.T) {
	laddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		t.Error(err)
	}
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		t.Error(err)
	}
	listenPort := listener.Addr().(*net.TCPAddr).Port

	// Start TCP Server
	server := NewBannerServer()
	go server.Serve(listener)

	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chPzATCP",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"port":%d,"ssl":false},
	  "type":"remote.tcp",
	  "timeout":15,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"",
	  "disabled":false
	  }`, listenPort)
	check := check.NewCheck([]byte(checkData), context.Background(), func() {})

	// Run check
	crs, err := check.Run()
	if err != nil {
		t.Error(err)
	}

	// Shutdown server
	server.Stop()

	// Validate Metrics
	if crs.Status != "success" {
		t.Fatal("status is not `success`")
	}

	ValidateMetrics(t, []string{"duration", "tt_connect"}, crs.Get(0))
}

func TestTCPRunFailureClosedPort(t *testing.T) {
	// Generate an unused port
	laddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		t.Error(err)
	}
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		t.Error(err)
	}
	listenPort := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chPzATCP",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"port":%d,"ssl":false},
	  "type":"remote.tcp",
	  "timeout":1,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"",
	  "disabled":false
	  }`, listenPort)
	check := check.NewCheck([]byte(checkData), context.Background(), func() {})

	// Run check
	crs, err := check.Run()
	if err != nil {
		t.Error(err)
	}

	// Validate Metrics
	//   - will be unavailable
	if crs.Available == true {
		t.Fatal("status must be not success")
	}

	if crs.Length() != 0 {
		t.Fatal("metric length should be 0")
	}
}
