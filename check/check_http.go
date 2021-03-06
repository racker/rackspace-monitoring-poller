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
	"math"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
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
func NewHTTPCheck(base *Base) (Check, error) {
	check := &HTTPCheck{Base: *base}
	err := json.Unmarshal(*base.RawDetails, &check.Details)
	if err != nil {
		log.WithFields(log.Fields{
			"prefix":  "check_http",
			"err":     err,
			"details": string(*base.RawDetails),
		}).Error("Unable to unmarshal check details")
		return nil, err
	}
	return check, nil
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

	var netClient *http.Client
	tlsConfig := &tls.Config{InsecureSkipVerify: true, ServerName: host}
	transport := &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig:   tlsConfig,
		DialContext:       NewCustomDialContext(ch.TargetResolver),
	}
	netClient = &http.Client{Transport: transport}

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
	}).Info("Running check")

	// Setup Request
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		crs.SetStatusFromError(err)
		crs.SetStateUnavailable()
		return crs, nil
	}

	// Setup Auth
	if len(ch.Details.AuthUser) > 0 && len(ch.Details.AuthPassword) > 0 {
		req.SetBasicAuth(ch.Details.AuthUser, ch.Details.AuthPassword)
	}
	// Setup Tracing
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
	req.Header.Add("User-Agent", UserAgent)
	req.Host = host
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
		crs.SetStatusFromError(err)
		crs.SetStateUnavailable()
		return crs, nil
	}
	endtime := utils.NowTimestampMillis()

	// Parse Body
	if len(ch.Details.Body) > 0 {
		re, err := regexp.Compile(ch.Details.Body)
		if err != nil {
			crs.SetStatusFromError(err)
			crs.SetStateUnavailable()
			return crs, nil
		}
		if m := re.FindSubmatch(body); m != nil {
			cr.AddMetric(metric.NewMetric("body_match", "", metric.MetricString, string(m[0]), "string"))
		} else {
			cr.AddMetric(metric.NewMetric("body_match", "", metric.MetricString, "", "string"))
		}
	}

	// Body Matches
	for key, regex := range ch.Details.BodyMatches {
		re, err := regexp.Compile(regex)
		if err != nil {
			crs.SetStatusFromError(err)
			crs.SetStateUnavailable()
			return crs, nil
		}
		if m := re.FindSubmatch(body); m != nil {
			cr.AddMetric(metric.NewMetric(fmt.Sprintf("body_match_%s", key), "", metric.MetricString, string(m[1]), ""))
		} else {
			cr.AddMetric(metric.NewMetric(fmt.Sprintf("body_match_%s", key), "", metric.MetricString, "", ""))
		}
	}

	truncated := int64(0)
	if int64(len(body)) > MaxHTTPResponseBodyLength {
		truncated = int64(1)
	}

	// Status Code Numeral
	var code100 int
	var code200 int
	var code300 int
	var code400 int
	var code500 int
	codeStr := strconv.Itoa(resp.StatusCode)
	codeFamily := int(math.Floor(float64(resp.StatusCode) / 100.0))
	switch codeFamily {
	case 1:
		code100 = 1
	case 2:
		code200 = 1
	case 3:
		code300 = 1
	case 4:
		code400 = 1
	case 5:
		code500 = 1
	}
	cr.AddMetric(metric.NewMetric("code_100", "", metric.MetricNumber, code100, ""))
	cr.AddMetric(metric.NewMetric("code_200", "", metric.MetricNumber, code200, ""))
	cr.AddMetric(metric.NewMetric("code_300", "", metric.MetricNumber, code300, ""))
	cr.AddMetric(metric.NewMetric("code_400", "", metric.MetricNumber, code400, ""))
	cr.AddMetric(metric.NewMetric("code_500", "", metric.MetricNumber, code500, ""))

	cr.AddMetric(metric.NewMetric("code", "", metric.MetricString, codeStr, ""))
	cr.AddMetric(metric.NewMetric("duration", "", metric.MetricNumber, endtime-starttime, "milliseconds"))
	cr.AddMetric(metric.NewMetric("bytes", "", metric.MetricNumber, len(body), "bytes"))
	cr.AddMetric(metric.NewMetric("truncated", "", metric.MetricNumber, truncated, "bool"))

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
