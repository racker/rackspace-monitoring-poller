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
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"regexp"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	protocol "github.com/racker/rackspace-monitoring-poller/protocol/check"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"
)

const (
	// MaxTCPBannerLength sets the maximum tcp banner lines
	MaxTCPBannerLength = int(80)
	// MaxTCPBodyLength sets the maximum tcp request body
	MaxTCPBodyLength = int64(1024)
)

// TCPCheck conveys TCP checks
type TCPCheck struct {
	Base
	protocol.TCPCheckDetails
}

// NewTCPCheck - Constructor for an TCP Check
func NewTCPCheck(base *Base) Check {
	check := &TCPCheck{Base: *base}
	err := json.Unmarshal(*base.RawDetails, &check.Details)
	if err != nil {
		log.Error("Error unmarshalling TCPCheck")
		return nil
	}
	return check
}

// GenerateAddress function creates an address
// from check port and target ip
func (ch *TCPCheck) GenerateAddress() (string, error) {
	portStr := strconv.FormatUint(ch.Details.Port, 10)
	ip, err := ch.GetTargetIP()
	if err != nil {
		return "", err
	}
	return net.JoinHostPort(ip, portStr), nil
}

func (ch *TCPCheck) readLine(conn io.Reader) ([]byte, error) {
	bio := bufio.NewReader(conn)
	line, _, err := bio.ReadLine()
	if err != nil {
		return nil, err
	}
	return line, nil
}

func calculateTimeout(dialer *net.Dialer) time.Duration {
	timeout := dialer.Timeout

	if !dialer.Deadline.IsZero() {
		deadlineTimeout := dialer.Deadline.Sub(time.Now())
		if timeout == 0 || deadlineTimeout < timeout {
			timeout = deadlineTimeout
		}
	}

	return timeout
}

func dialContextWithDialer(ctx context.Context, dialer *net.Dialer, network, addr string, config *tls.Config) (net.Conn, error) {
	timeout := calculateTimeout(dialer)
	if timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	rawConn, err := dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}
	// tls disabled
	if config == nil {
		return rawConn, nil
	}

	conn := tls.Client(rawConn, config)
	errChannel := make(chan error, 1)

	go func() {
		errChannel <- conn.Handshake()
	}()

	select {
	case err := <-errChannel:
		if err != nil {
			rawConn.Close()
			return nil, err
		}
		return conn, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Run method implements Check.Run method for TCP
// please see Check interface for more information
func (ch *TCPCheck) Run() (*ResultSet, error) {
	// TODO: refactor - cyclomatic complexity is high (16)
	var conn net.Conn
	var err error
	var endtime int64

	cr := NewResult()
	crs := NewResultSet(ch, cr)
	starttime := utils.NowTimestampMillis()
	addr, _ := ch.GenerateAddress()
	log.WithFields(log.Fields{
		"prefix":  ch.GetLogPrefix(),
		"address": addr,
		"ssl":     ch.Details.UseSSL,
	}).Info("Running TCP Check")

	ctx := context.Background()
	timeout := ch.GetTimeoutDuration()
	nd := &net.Dialer{

		Timeout: timeout,
	}

	network := "tcp4"
	if ch.TargetResolver == ResolverIPV6 {
		network = "tcp6"
	}

	// Connection
	if ch.Details.UseSSL {
		TLSconfig := &tls.Config{InsecureSkipVerify: true}
		conn, err = dialContextWithDialer(ctx, nd, network, addr, TLSconfig)
	} else {
		conn, err = dialContextWithDialer(ctx, nd, network, addr, nil)
	}
	if err != nil {
		crs.SetStatus(err.Error())
		crs.SetStateUnavailable()
		return crs, nil
	}
	defer conn.Close()

	connectEndTime := utils.NowTimestampMillis()
	cr.AddMetric(metric.NewMetric("tt_connect", "", metric.MetricNumber, connectEndTime-starttime, metric.UnitMilliseconds))

	// Set read/write timeout
	conn.SetDeadline(time.Now().Add(time.Duration(ch.GetTimeout()) * time.Millisecond))

	// Send Body
	if len(ch.Details.SendBody) > 0 {
		// TODO: this can throw an exception and stop the flow
		// do we want to test for error and log/continue or break
		// and return HTTP 500?
		_, err := io.WriteString(conn, ch.Details.SendBody)
		if err != nil {
			crs.SetStatus(err.Error())
			crs.SetStateUnavailable()
			return crs, nil
		}
	}

	// Banner Match
	if len(ch.Details.BannerMatch) > 0 {
		line, err := ch.readLine(conn)
		if err != nil {
			crs.SetStatus(err.Error())
			crs.SetStateUnavailable()
			return crs, nil
		}
		firstbytetime := utils.NowTimestampMillis()
		// return a fixed size banner
		if len(line) > MaxTCPBannerLength {
			line = line[:MaxTCPBannerLength]
		}
		if re, err := regexp.Compile(ch.Details.BannerMatch); err == nil {
			if m := re.FindSubmatch(line); m != nil {
				cr.AddMetric(metric.NewMetric("banner_match", "", metric.MetricString, string(m[1]), ""))
			} else {
				cr.AddMetric(metric.NewMetric("banner_match", "", metric.MetricString, "", ""))
			}
		}
		cr.AddMetric(metric.NewMetric("tt_firstbyte", "", metric.MetricNumber, firstbytetime-starttime, metric.UnitMilliseconds))
	}

	// Body Match
	if len(ch.Details.BodyMatch) > 0 {
		body, err := ch.readLimit(conn, MaxTCPBodyLength)
		if err != nil {
			crs.SetStatus(err.Error())
			crs.SetStateUnavailable()
			return crs, nil
		}
		bodybytetime := utils.NowTimestampMillis()
		cr.AddMetric(metric.NewMetric("tt_body", "", metric.MetricNumber, bodybytetime-starttime, metric.UnitMilliseconds))
		if re, err := regexp.Compile(ch.Details.BodyMatch); err != nil {
			if m := re.FindAllStringSubmatch(string(body), -1); m != nil {
				for _, s := range m {
					if len(s) == 2 {
						cr.AddMetric(metric.NewMetric(s[0], "", metric.MetricString, s[1], ""))
					}
				}

			} else {
				cr.AddMetric(metric.NewMetric("body_match", "", metric.MetricString, "", ""))
			}
		}
	}
	endtime = utils.NowTimestampMillis()
	cr.AddMetric(metric.NewMetric("duration", "", metric.MetricNumber, endtime-starttime, metric.UnitMilliseconds))

	// TLS Metrics
	if ch.Details.UseSSL {
		tlsConn := conn.(*tls.Conn)
		ch.AddTLSMetrics(cr, tlsConn.ConnectionState())
	}

	crs.SetStateAvailable()
	crs.SetStatusSuccess()
	return crs, nil
}
