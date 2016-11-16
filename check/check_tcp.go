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
	"crypto/tls"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"io"
	"net"
	"regexp"
	"strconv"
	"time"
)

const (
	MaxTCPBannerLength = int(80)
	MaxTCPBodyLength   = int64(1024)
)

type TCPCheck struct {
	CheckBase
	Details struct {
		BannerMatch string `json:"banner_match"`
		BodyMatch   string `json:"body_match"`
		Port        uint64 `json:"port"`
		SendBody    string `json:"send_body"`
		UseSSL      bool   `json:"ssl"`
	}
}

func NewTCPCheck(base *CheckBase) Check {
	check := &TCPCheck{CheckBase: *base}
	err := json.Unmarshal(*base.Details, &check.Details)
	if err != nil {
		log.Error("Error unmarshalling TCPCheck")
		return nil
	}
	check.PrintDefaults()
	return check
}

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

func (ch *TCPCheck) readLimit(conn io.Reader, limit int64) ([]byte, error) {
	bytes := make([]byte, limit)
	bio := io.LimitReader(conn, limit)
	count, err := bio.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes[:count], nil
}

func (ch *TCPCheck) Run() (*CheckResultSet, error) {
	var conn net.Conn
	var err error
	var endtime int64

	cr := NewCheckResult()
	crs := NewCheckResultSet(ch, cr)
	crs.SetStateUnavailable()
	crs.SetStatusUnknown()
	starttime := utils.NowTimestampMillis()
	addr, _ := ch.GenerateAddress()
	log.WithFields(log.Fields{
		"type":    ch.CheckType,
		"id":      ch.Id,
		"address": addr,
		"ssl":     ch.Details.UseSSL,
	}).Info("Running TCP Check")

	// Connection
	nd := &net.Dialer{Timeout: time.Duration(ch.GetTimeout()) * time.Millisecond}
	if ch.Details.UseSSL {
		TLSconfig := &tls.Config{}
		conn, err = tls.DialWithDialer(nd, "tcp", addr, TLSconfig)
	} else {
		conn, err = nd.Dial("tcp", addr)
	}
	if err != nil {
		crs.SetStatus(err.Error())
		crs.SetStateUnavailable()
		return crs, nil
	}
	defer conn.Close()

	// Set read/write timeout
	conn.SetDeadline(time.Now().Add(time.Duration(ch.GetTimeout()) * time.Millisecond))

	// Send Body
	if len(ch.Details.SendBody) > 0 {
		io.WriteString(conn, ch.Details.SendBody)
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
		if re, err := regexp.Compile(ch.Details.BannerMatch); err != nil {
			if m := re.FindSubmatch(line); m != nil {
				cr.AddMetric(metric.NewMetric("banner_match", "", metric.MetricString, m[0], ""))
			} else {
				cr.AddMetric(metric.NewMetric("banner_match", "", metric.MetricString, "", ""))
			}
		}
		cr.AddMetric(metric.NewMetric("tt_firstbyte", "", metric.MetricNumber, firstbytetime-starttime, "ms"))
	}

	// Body Match
	if len(ch.Details.BodyMatch) > 0 {
		body, err := ch.readLimit(conn, MaxTCPBodyLength)
		if err != nil {
			return crs, nil
		}
		bodybytetime := utils.NowTimestampMillis()
		cr.AddMetric(metric.NewMetric("tt_body", "", metric.MetricNumber, bodybytetime-starttime, "ms"))
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
	cr.AddMetric(metric.NewMetric("duration", "", metric.MetricNumber, endtime-starttime, "ms"))
	crs.SetStateAvailable()
	crs.SetStatusSuccess()
	return crs, nil
}
