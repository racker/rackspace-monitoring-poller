package check

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/metric"
	ping "github.com/sparrc/go-ping"
	"time"
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

	pinger.Count = int(ch.Details.Count)
	pinger.Timeout = time.Duration(ch.Timeout) * time.Second

	pinger.OnRecv = func(pkt *ping.Packet) {
		log.WithFields(log.Fields{
			"id":    ch.GetId(),
			"bytes": pkt.Nbytes,
			"seq":   pkt.Seq,
			"rtt":   pkt.Rtt,
		}).Debug("Received ping packet")
	}

	log.WithFields(log.Fields{
		"id":         ch.GetId(),
		"count":      pinger.Count,
		"timeoutSec":  ch.Timeout,
		"timeoutDur": pinger.Timeout,
	}).Debug("Starting pinger")

	// blocking
	pinger.Run()

	stats := pinger.Statistics()

	cr := NewCheckResult(
		metric.NewMetric("available", "", metric.MetricFloat, stats.PacketsRecv/stats.PacketsSent, ""),
		metric.NewMetric("average", "", metric.MetricFloat, float64(stats.AvgRtt/time.Millisecond), "ms"),
		metric.NewMetric("count", "", metric.MetricNumber, stats.PacketsSent, ""),
		metric.NewMetric("maximum", "", metric.MetricFloat, float64(stats.MaxRtt/time.Millisecond), "ms"),
		metric.NewMetric("minimum", "", metric.MetricFloat, float64(stats.MinRtt/time.Millisecond), "ms"),
	)

	log.WithFields(log.Fields{
		"id":     ch.GetId(),
		"stats":  stats,
		"result": cr,
	}).Debug("Finished remote.ping check")

	return NewCheckResultSet(ch, cr), nil
}
