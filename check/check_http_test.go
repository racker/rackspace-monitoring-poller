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
	check "github.com/racker/rackspace-monitoring-poller/check"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	staticHello = "Hello, world"
)

func staticResponse(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "##############")
	for key, values := range r.Header {
		fmt.Fprintln(w, fmt.Sprintf("%v=%v", key, strings.Join(values, ",")))
	}
	fmt.Fprintln(w, "##############")
	fmt.Fprintln(w, staticHello)
}

func TestHTTPSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(staticResponse))
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
		"truncated",
		"tt_connect",
		"tt_firstbyte",
	}
	ValidateMetrics(t, metrics, crs.Get(0))
}

func TestHTTPSuccessIncludeBodyAndHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(staticResponse))
	defer ts.Close()

	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chPzAHTTP",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"url":"%s","include_body":true,"headers":{"foo":"bar"}},
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
		"body",
		"truncated",
		"tt_connect",
		"tt_firstbyte",
	}
	ValidateMetrics(t, metrics, crs.Get(0))
	// Validate body
	body, _ := crs.Get(0).GetMetric("body").ToString()
	if !strings.Contains(body, staticHello) {
		t.Fatal("body does not contain: " + staticHello)
	}

	if !strings.Contains(body, "Foo=bar") {
		t.Fatal("header is not present")
	}
}

func TestHTTPSuccessBodyMatch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(staticResponse))
	defer ts.Close()

	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chPzAHTTP",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"url":"%s","include_body":true,"headers":{"foo":"bar"},"body":"Foo=(.*)","body_matches":{"foo":"Foo=(.*)"}},
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
		"body",
		"body_match",
		"body_match_foo",
		"truncated",
		"tt_connect",
		"tt_firstbyte",
	}
	ValidateMetrics(t, metrics, crs.Get(0))

	// Validate body
	body, _ := crs.Get(0).GetMetric("body").ToString()
	if !strings.Contains(body, staticHello) {
		t.Fatal("body does not contain: " + staticHello)
	}
	if !strings.Contains(body, "Foo=bar") {
		t.Fatal("header is not present")
	}
	// Validate body_match
	bodyMatch, _ := crs.Get(0).GetMetric("body_match").ToString()
	if !strings.Contains(bodyMatch, "bar") {
		t.Fatal("bodyMatch does not contain bar")
	}
	// Validate body_match_foo
	foo, _ := crs.Get(0).GetMetric("body_match_foo").ToString()
	if !strings.Contains(foo, "bar") {
		t.Fatal("foo does not contain bar")
	}
}

func TestHTTPClosed(t *testing.T) {
	// Create a server then close it
	ts := httptest.NewServer(http.HandlerFunc(staticResponse))
	ts.Close()

	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chPzAHTTP",
	  "zone_id":"pzA",
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
	if crs.Available == true {
		t.Fatal("availability should be false")
	}
}

func TestHTTPTimeout(t *testing.T) {
	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Create a server that times out
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-ctx.Done()
	}))
	defer ts.Close()

	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chPzAHTTPTimeout",
	  "zone_id":"pzA",
	  "details":{"url":"%s"},
	  "type":"remote.http",
	  "timeout":1,
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
	if crs.Available == true {
		t.Fatal("availability should be false")
	}

	cancel()
}
