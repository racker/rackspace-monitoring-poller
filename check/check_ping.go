package check

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"encoding/json"
	ping "github.com/sparrc/go-ping"
	"time"
	"github.com/racker/rackspace-monitoring-poller/metric"
)

type PingCheck struct {
	CheckBase
	Details struct {
		Count uint8 `json:"count"`
	}
}

func NewPingCheck(base *CheckBase) Check {
	check := &PingCheck{CheckBase: *base}
	err := json.Unmarshal(*base.Details, &check.Details)
	if err != nil {
		log.Printf("Error unmarshalling check details")
		return nil
	}
	check.PrintDefaults()
	return check
}

func (ch *PingCheck) Run() (*CheckResultSet, error) {
	log.Printf("Running PING Check: %v", ch.GetId())

	targetIp, err := ch.GetTargetIP()
	if err != nil {
		return nil, err
	}

	pinger, err := ping.NewPinger(targetIp)
	if err != nil {
		log.WithField("targetIp", targetIp).
			Error("Failed to create pinger")
		return nil, err
	}

	pinger.Count = ch.Details.Count
	pinger.Timeout = ch.Timeout * time.Millisecond

	pinger.OnRecv = func(pkt *ping.Packet) {
		log.WithFields(log.Fields{
				"bytes": pkt.Nbytes,
				"seq": pkt.Seq,
				"rtt": pkt.Rtt,
			}).Debug("Received ping packet")
	}

	// blocking
	pinger.Run()

	stats := pinger.Statistics()

	cr := NewCheckResult(
		metric.NewMetric("available", "", metric.MetricFloat, stats.PacketsRecv/stats.PacketsSent, ""),
		metric.NewMetric("average", "", metric.MetricFloat, stats.AvgRtt / time.Millisecond, "ms"),
		metric.NewMetric("count", "", metric.MetricNumber, stats.PacketsSent, ""),
		metric.NewMetric("maximum", "", metric.MetricFloat, stats.MaxRtt / time.Millisecond, "ms"),
		metric.NewMetric("minimum", "", metric.MetricFloat, stats.MinRtt / time.Millisecond, "ms"),
	)

	return NewCheckResultSet(ch, cr), nil
}
