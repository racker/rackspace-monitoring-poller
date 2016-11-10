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

var (
	MaxTCPBannerLength = int(80)
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
	starttime := utils.NowTimestampMillis()
	addr, _ := ch.GenerateAddress()
	log.WithFields(log.Fields{
		"type":    ch.CheckType,
		"id":      ch.Id,
		"address": addr,
		"ssl":     ch.Details.UseSSL,
	}).Info("Running TCP Check")
	// Connection
	nd := &net.Dialer{Timeout: time.Duration(ch.GetTimeout()) * time.Second}
	if ch.Details.UseSSL {
		TLSconfig := &tls.Config{}
		conn, err = tls.DialWithDialer(nd, "tcp", addr, TLSconfig)
	} else {
		conn, err = nd.Dial("tcp", addr)
	}
	if err != nil {
		log.Error(err)
		return crs, nil
	}
	defer conn.Close()

	// Set read/write timeout
	conn.SetDeadline(time.Now().Add(time.Duration(ch.GetTimeout()) * time.Second))

	// Send Body
	if ch.Details.SendBody != "" {
		io.WriteString(conn, ch.Details.SendBody)
	}

	// Banner Match
	if len(ch.Details.BannerMatch) > 0 {
		line, err := ch.readLine(conn)
		if err != nil {
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
		body, err := ch.readLimit(conn, 1024)
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
	return crs, nil
}
