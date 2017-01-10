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

// HTTP Check
package check

import (
	"context"
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	protocol "github.com/racker/rackspace-monitoring-poller/protocol/check"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	// MaxHttpResponseBodyLength Maxiumum Allowed Body Length
	MaxHttpResponseBodyLength = int64(512 * 1024)
	// UserAgent the header value to send for the user agent
	UserAgent = "Rackspace Monitoring Poller/1.0 (https://monitoring.api.rackspacecloud.com/)"
)

// HTTPCheck conveys HTTP checks
type HTTPCheck struct {
	CheckBase
	protocol.HTTPCheckDetails
}

// Constructor for an HTTP Check
func NewHTTPCheck(base *CheckBase) Check {
	check := &HTTPCheck{CheckBase: *base}
	err := json.Unmarshal(*base.RawDetails, &check.Details)
	if err != nil {
		log.Error("Error unmarshalling checkbase")
		return nil
	}
	check.PrintDefaults()
	return check
}

func (ch *HTTPCheck) parseTLS(cr *CheckResult, resp *http.Response) {
	cert := resp.TLS.PeerCertificates[0]
	cr.AddMetric(metric.NewMetric("cert_serial", "", metric.MetricNumber, cert.SerialNumber, ""))
	if len(cert.OCSPServer) > 0 {
		cr.AddMetric(metric.NewMetric("cert_ocsp", "", metric.MetricNumber, cert.OCSPServer[0], ""))
	}
	switch cert.PublicKeyAlgorithm {
	case x509.RSA:
		publicKey := cert.PublicKey.(*rsa.PublicKey)
		cr.AddMetric(metric.NewMetric("cert_bits", "", metric.MetricNumber, publicKey.N.BitLen(), ""))
		cr.AddMetric(metric.NewMetric("cert_type", "", metric.MetricNumber, "rsa", ""))
	case x509.DSA:
		publicKey := cert.PublicKey.(*dsa.PublicKey)
		cr.AddMetric(metric.NewMetric("cert_bits", "", metric.MetricNumber, publicKey.Q.BitLen(), ""))
		cr.AddMetric(metric.NewMetric("cert_type", "", metric.MetricNumber, "dsa", ""))
	case x509.ECDSA:
		publicKey := cert.PublicKey.(*ecdsa.PublicKey)
		cr.AddMetric(metric.NewMetric("cert_bits", "", metric.MetricNumber, publicKey.Params().BitSize, ""))
		cr.AddMetric(metric.NewMetric("cert_type", "", metric.MetricNumber, "ecdsa", ""))
	default:
		cr.AddMetric(metric.NewMetric("cert_bits", "", metric.MetricNumber, "0", ""))
		cr.AddMetric(metric.NewMetric("cert_type", "", metric.MetricNumber, "-", ""))
	}
	cr.AddMetric(metric.NewMetric("cert_sig_algo", "", metric.MetricNumber, strings.ToLower(cert.SignatureAlgorithm.String()), ""))
	var sslVersion string
	switch resp.TLS.Version {
	case tls.VersionSSL30:
		sslVersion = "ssl3"
	case tls.VersionTLS10:
		sslVersion = "tls1.0"
	case tls.VersionTLS11:
		sslVersion = "tls1.1"
	case tls.VersionTLS12:
		sslVersion = "tls1.2"
	}
	cr.AddMetric(metric.NewMetric("ssl_session_version", "", metric.MetricNumber, sslVersion, ""))
	var cipherSuite string
	switch resp.TLS.CipherSuite {

	case tls.TLS_RSA_WITH_RC4_128_SHA:
		cipherSuite = "TLS_RSA_WITH_RC4_128_SHA"
	case tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA:
		cipherSuite = "TLS_RSA_WITH_3DES_EDE_CBC_SHA"
	case tls.TLS_RSA_WITH_AES_128_CBC_SHA:
		cipherSuite = "TLS_RSA_WITH_AES_128_CBC_SHA"
	case tls.TLS_RSA_WITH_AES_256_CBC_SHA:
		cipherSuite = "TLS_RSA_WITH_AES_256_CBC_SHA"
	case tls.TLS_RSA_WITH_AES_128_GCM_SHA256:
		cipherSuite = "TLS_RSA_WITH_AES_128_GCM_SHA256"
	case tls.TLS_RSA_WITH_AES_256_GCM_SHA384:
		cipherSuite = "TLS_RSA_WITH_AES_256_GCM_SHA384"
	case tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA:
		cipherSuite = "TLS_ECDHE_ECDSA_WITH_RC4_128_SHA"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA:
		cipherSuite = "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:
		cipherSuite = "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA"
	case tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA:
		cipherSuite = "TLS_ECDHE_RSA_WITH_RC4_128_SHA"
	case tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA:
		cipherSuite = "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA"
	case tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA:
		cipherSuite = "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA"
	case tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:
		cipherSuite = "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA"
	case tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:
		cipherSuite = "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:
		cipherSuite = "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"
	case tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:
		cipherSuite = "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:
		cipherSuite = "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"
	default:
		cipherSuite = "-"
	}
	cr.AddMetric(metric.NewMetric("ssl_session_cipher", "", metric.MetricNumber, cipherSuite, ""))
	//TODO Fill in the rest of the SSL info
}

func disableRedirects(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

func (ch *HTTPCheck) Run() (*CheckResultSet, error) {
	log.WithFields(log.Fields{
		"type": ch.CheckType,
		"id":   ch.Id,
	}).Info("Running HTTP Check")

	ctx, cancel := context.WithTimeout(context.Background(), ch.GetTimeoutDuration())
	defer cancel()

	sl := utils.NewStatusLine()
	cr := NewCheckResult()
	crs := NewCheckResultSet(ch, cr)
	starttime := utils.NowTimestampMillis()

	// Parse URL and Replace Host with IP
	parsed, err := url.Parse(ch.Details.Url)
	if err != nil {
		return nil, err
	}

	host, port, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		return nil, err
	}
	ip, err := ch.GetTargetIP()
	if err != nil {
		return nil, err
	}
	if len(port) > 0 {
		parsed.Host = net.JoinHostPort(ip, port)
	} else {
		parsed.Host = ip
	}
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
	body, err := ch.readLimit(resp.Body, MaxHttpResponseBodyLength)
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
	codeStr := strconv.FormatInt(int64(resp.StatusCode), 10)

	cr.AddMetric(metric.NewMetric("code", "", metric.MetricString, codeStr, ""))
	cr.AddMetric(metric.NewMetric("duration", "", metric.MetricNumber, endtime-starttime, "milliseconds"))
	cr.AddMetric(metric.NewMetric("bytes", "", metric.MetricNumber, len(body), "bytes"))
	cr.AddMetric(metric.NewMetric("truncated", "", metric.MetricNumber, truncated, "bytes"))

	if ch.Details.IncludeBody {
		cr.AddMetric(metric.NewMetric("body", "", metric.MetricString, string(body), ""))
	}

	// TLS
	if resp.TLS != nil {
		ch.parseTLS(cr, resp)
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
