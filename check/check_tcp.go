package check

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	log "github.com/Sirupsen/logrus"

	"io"
	"net"
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

func (ch *TCPCheck) Run() (*CheckResultSet, error) {
	var conn net.Conn
	var err error
	cr := NewCheckResult()
	starttime := NowTimestampMillis()
	addr, _ := ch.GenerateAddress()
	nd := &net.Dialer{Timeout: time.Duration(ch.GetTimeout()) * time.Second}
	log.WithFields(log.Fields{
		"type":    ch.CheckType,
		"id":      ch.Id,
		"address": addr,
		"ssl":     ch.Details.UseSSL,
	}).Info("Running TCP Check")
	if ch.Details.UseSSL {
		TLSconfig := &tls.Config{}
		conn, err = tls.DialWithDialer(nd, "tcp", addr, TLSconfig)
	} else {
		conn, err = nd.Dial("tcp", addr)
	}
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(time.Duration(ch.GetTimeout()) * time.Second))
	if ch.Details.SendBody != "" {
		io.WriteString(conn, ch.Details.SendBody)
	}
	if ch.Details.BannerMatch != "" {
		bio := bufio.NewReader(conn)
		line, _, err := bio.ReadLine()
		if err != nil {
			log.Error(err)
			return nil, err
		}
		firstbytetime := NowTimestampMillis()
		if len(line) > MaxTCPBannerLength {
			line = line[:MaxTCPBannerLength]
		}
		cr.AddMetric(NewMetric("tt_firstbyte", "", MetricNumber, firstbytetime-starttime, "ms"))
		cr.AddMetric(NewMetric("banner", "", MetricString, string(line), ""))
	}
	endtime := NowTimestampMillis()
	cr.AddMetric(NewMetric("duration", "", MetricNumber, endtime-starttime, "ms"))
	return NewCheckResultSet(ch, cr), nil
}
