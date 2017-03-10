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

package check

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	protocol "github.com/racker/rackspace-monitoring-poller/protocol/check"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"
)

var (
	// MaxHTTPResponseBodyLength Maxiumum Allowed Body Length
	MaxHTTPResponseBodyLength = int64(512 * 1024)
	// UserAgent the header value to send for the user agent
	UserAgent = "Rackspace Monitoring Poller/1.0 (https://monitoring.api.rackspacecloud.com/)"
	// DefaultPort the default http port is 80
	DefaultPort = "80"
	// DefaultSecurePort the default TLS port is 443
	DefaultSecurePort = "443"
)

// HTTPCheck conveys HTTP checks
type HTTPCheck struct {
	Base
	protocol.HTTPCheckDetails
}

// NewHTTPCheck - Constructor for an HTTP Check
func NewHTTPCheck(base *Base) Check {
	check := &HTTPCheck{Base: *base}
	err := json.Unmarshal(*base.RawDetails, &check.Details)
	if err != nil {
		log.Error("Error unmarshalling base")
		return nil
	}
	return check
}

func disableRedirects(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

// Run method implements Check.Run method for HTTP
// please see Check interface for more information
func (ch *HTTPCheck) Run() (*ResultSet, error) {
	// TODO: refactor.  High cyclomatic complexity (21)
	log.WithFields(log.Fields{
		"prefix": ch.GetLogPrefix(),
		"type":   ch.CheckType,
		"id":     ch.Id,
	}).Debug("Running HTTP Check")

	ctx, cancel := context.WithTimeout(context.Background(), ch.GetTimeoutDuration())
	defer cancel()

	sl := utils.NewStatusLine()
	cr := NewResult()
	crs := NewResultSet(ch, cr)
	starttime := utils.NowTimestampMillis()

	// Parse URL and Replace Host with IP
	parsed, err := url.Parse(ch.Details.Url)
	if err != nil {
		return nil, err
	}

	host, port, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			if len(port) == 0 {
				if parsed.Scheme == "http" {
					port = DefaultPort
				} else {
					port = DefaultSecurePort
				}
				host = parsed.Host
			}
		} else {
			return nil, err
		}
	}
	ip, err := ch.GetTargetIP()
	if err != nil && err != ErrInvalidTargetIP {
		return crs, err
	}
	if ip == "" {
		log.WithFields(log.Fields{
			"prefix": ch.GetLogPrefix(),
		}).Debug("Setting host to IP. IP was an empty string")
		ip = host
	}
	parsed.Host = net.JoinHostPort(ip, port)
	url := parsed.String()

	// Setup HTTP or HTTPS Client
	var netClient *http.Client
	if parsed.Scheme == "http" {
		netClient = &http.Client{}
	} else {
		tlsConfig := &tls.Config{InsecureSkipVerify: true, ServerName: host}
		transport := &http.Transport{TLSClientConfig: tlsConfig}
		netClient = &http.Client{Transport: transport}
	}

	// Setup Redirects
	if !ch.Details.FollowRedirects {
		netClient.CheckRedirect = disableRedirects
	}

	// Setup Method
	method := strings.ToUpper(ch.Details.Method)

	log.WithFields(log.Fields{
		"prefix": ch.GetLogPrefix(),
		"method": method,
		"url":    url,
	}).Debug("Debug Request")

	// Setup Request
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		crs.SetStatus(err.Error())
		crs.SetStateUnavailable()
		return crs, nil
	}
	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			firstbytetime := utils.NowTimestampMillis()
			cr.AddMetric(metric.NewMetric("tt_firstbyte", "", metric.MetricNumber, firstbytetime-starttime, "milliseconds"))
		},
		ConnectDone: func(network, addr string, err error) {
			connectdonetime := utils.NowTimestampMillis()
			cr.AddMetric(metric.NewMetric("tt_connect", "", metric.MetricNumber, connectdonetime-starttime, "milliseconds"))

		},
	}
	req = req.WithContext(httptrace.WithClientTrace(ctx, trace))

	// Add Headers
	req.Header.Add("Accept-Encoding", "gzip,deflate")
	req.Header.Add("User-Agent", UserAgent)
	req.Header.Add("Host", host)
	for key, value := range ch.Details.Headers {
		req.Header.Add(key, value)
	}

	// Perform Request
	resp, err := netClient.Do(req)
	if err != nil {
		crs.SetStatus("connection refused")
		crs.SetStateUnavailable()
		return crs, nil
	}
	defer resp.Body.Close()

	// Read Body
	body, err := ch.readLimit(resp.Body, MaxHTTPResponseBodyLength)
	if err != nil {
		crs.SetStatus(err.Error())
		crs.SetStateUnavailable()
		return crs, nil
	}
	endtime := utils.NowTimestampMillis()

	// Parse Body
	if len(ch.Details.Body) > 0 {
		re, err := regexp.Compile(ch.Details.Body)
		if err != nil {
			crs.SetStatus(err.Error())
			crs.SetStateUnavailable()
			return crs, nil
		}
		if m := re.FindSubmatch(body); m != nil {
			cr.AddMetric(metric.NewMetric("body_match", "", metric.MetricString, string(m[1]), ""))
		} else {
			cr.AddMetric(metric.NewMetric("body_match", "", metric.MetricString, "", ""))
		}
	}

	// Body Matches
	for key, regex := range ch.Details.BodyMatches {
		re, err := regexp.Compile(regex)
		if err != nil {
			crs.SetStatus(err.Error())
			crs.SetStateUnavailable()
			return crs, nil
		}
		if m := re.FindSubmatch(body); m != nil {
			cr.AddMetric(metric.NewMetric(fmt.Sprintf("body_match_%s", key), "", metric.MetricString, string(m[1]), ""))
		} else {
			cr.AddMetric(metric.NewMetric(fmt.Sprintf("body_match_%s", key), "", metric.MetricString, "", ""))
		}
	}

	truncated := resp.ContentLength - int64(len(body))
	codeStr := strconv.Itoa(resp.StatusCode)

	cr.AddMetric(metric.NewMetric("code", "", metric.MetricString, codeStr, ""))
	cr.AddMetric(metric.NewMetric("duration", "", metric.MetricNumber, endtime-starttime, "milliseconds"))
	cr.AddMetric(metric.NewMetric("bytes", "", metric.MetricNumber, len(body), "bytes"))
	cr.AddMetric(metric.NewMetric("truncated", "", metric.MetricNumber, truncated, "bytes"))

	if ch.Details.IncludeBody {
		cr.AddMetric(metric.NewMetric("body", "", metric.MetricString, string(body), ""))
	}

	// TLS
	if resp.TLS != nil {
		metrics := ch.AddTLSMetrics(cr, *resp.TLS)
		if !metrics.Verified {
			sl.AddOption("sslerror")
		}
	}

	// Status Line
	sl.Add("code", resp.StatusCode)
	sl.Add("duration", endtime-starttime)
	sl.Add("bytes", len(body))
	sl.Add("truncated", truncated)

	crs.SetStateAvailable()
	crs.SetStatus(sl.String())
	return crs, nil
}
