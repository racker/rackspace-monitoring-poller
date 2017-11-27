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
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/stretchr/testify/require"
)

func ValidateMetrics(t *testing.T, metrics []string, cr *check.Result) {
	for _, metricName := range metrics {
		if metric := cr.GetMetric(metricName); metric == nil {
			log.Fatal("metric " + metricName + " does not exist")
		}
	}

}
// IsInRange returns true if the time xvalue is within buffer seconds of
// yvalue,
func IsInRange(xvalue int64, yvalue int64, buffer int) bool {
    delta := int(xvalue - yvalue)
    return -buffer < delta && delta < buffer
}

func TestTCP_TLSRunSuccess(t *testing.T) {
	cert, _ := tls.X509KeyPair(utils.LocalhostCert, utils.LocalhostKey)
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
	tlsListener, _ := tls.Listen("tcp", "127.0.0.1:0", tlsConfig)
	listenPort := tlsListener.Addr().(*net.TCPAddr).Port

	// Start TCP Server
	server := utils.NewBannerServer()
	go server.ServeTLS(tlsListener)

	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chTestTCP_TLSRunSuccess",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"port":%d,"ssl":true},
	  "type":"remote.tcp",
	  "timeout":15,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"IPv4",
	  "disabled":false
	  }`, listenPort)
	check, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	// Run check
	crs, err := check.Run()
	if err != nil {
		t.Error(err)
	}

	// Shutdown server
	server.Stop()
	tlsListener.Close()

	// Validate
	ValidateMetrics(t, []string{"duration", "tt_connect"}, crs.Get(0))
	cr := crs.Get(0)
	issuer, _ := cr.GetMetric("cert_issuer").ToString()
	if issuer != "/O=Acme Co" {
		t.Fatal("invalid issuer")
	}
	subject, _ := cr.GetMetric("cert_subject").ToString()
	if subject != "/O=Acme Co" {
		t.Fatal("invalid subject")
	}
	cert_start, _ := cr.GetMetric("cert_start").ToInt64()
	if cert_start != 0 {
		t.Fatal("invalid start time")
	}
	cert_end, _ := cr.GetMetric("cert_end").ToInt64()
	if cert_end != 3600000000 {
		t.Fatal("invalid end time")
	}
	cert_end_in, _ := cr.GetMetric("cert_end_in").ToInt64()
	expectedTime := cert_end - time.Now().Unix()
	// Since thie metric compares against current time, allow a 2s buffer
	if !IsInRange(cert_end_in, expectedTime, 2) {
		t.Fatal("invalid end in time")
	}
	cert_dns_names, _ := cr.GetMetric("cert_subject_alternate_names").ToString()
	if cert_dns_names != "example.com" {
		t.Fatal("invalid dns names")
	}
	cert_error, _ := cr.GetMetric("cert_error").ToString()
	if !strings.Contains(cert_error, "certificate signed by unknown authority") {
		t.Fatal("certificate should have unknown authority")
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
	server := utils.NewBannerServer()
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
	  "target_resolver":"IPv4",
	  "disabled":false
	  }`, listenPort)
	check, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	// Run check
	crs, err := check.Run()
	if err != nil {
		t.Error(err)
	}

	// Shutdown server
	server.Stop()
	listener.Close()

	// Validate Metrics
	if crs.Status != "success" {
		t.Fatal("status is not `success`")
	}

	ValidateMetrics(t, []string{"duration", "tt_connect"}, crs.Get(0))
}

func TestTCPRunFailureClosedPort(t *testing.T) {
	// TODO: This behavior has changed in 1.8.x versions of golang
	// TODO: Skipping for now
	t.Skip()

	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chPzATCP",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"port":1,"ssl":false},
	  "type":"remote.tcp",
	  "timeout":1,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"IPv4",
	  "disabled":false
	  }`)
	check, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	// Run check
	crs, err := check.Run()
	if err != nil {
		t.Error(err)
	}

	// Validate Metrics
	//   - will be unavailable
	if crs.Available {
		t.Fatal("status must be not success")
	}

	if crs.Length() != 0 {
		t.Fatal("metric length should be 0")
	}
}
