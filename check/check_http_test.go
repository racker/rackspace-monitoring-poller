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
	"fmt"
	check "github.com/racker/rackspace-monitoring-poller/check"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPRun(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chPzAHTTP",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"url":"%s"},
	  "type":"remote.http",
	  "timeout":15,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"",
	  "disabled":false
	  }`, ts.URL)
	check := check.NewCheck([]byte(checkData))

	// Run check
	crs, err := check.Run()
	if err != nil {
		t.Error(err)
	}

	// Validate Metrics
	if crs.Available == false {
		t.Fatal("availability should be true")
	}

	metrics := []string{
		"bytes",
		"code",
		"duration",
		"truncated",
		"tt_connect",
		"tt_firstbyte",
	}
	ValidateMetrics(t, metrics, crs.Get(0))
}
