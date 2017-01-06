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

// Ping Check
package check

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	protocol "github.com/racker/rackspace-monitoring-poller/protocol/check"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"
	ping "github.com/sparrc/go-ping"
	"time"
)

type PingCheck struct {
	CheckBase
	protocol.PingCheckDetails
}

func NewPingCheck(base *CheckBase) Check {
	check := &PingCheck{CheckBase: *base}
	err := json.Unmarshal(*base.RawDetails, &check.Details)
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

	pinger, err := PingerFactory(targetIp)
	if err != nil {
		log.WithField("targetIp", targetIp).
			Error("Failed to create pinger")
		return nil, err
	}

	pinger.SetCount(int(ch.Details.Count))
	pinger.SetTimeout(ch.GetTimeoutDuration())

	pinger.SetOnRecv(func(pkt *ping.Packet) {
		log.WithFields(log.Fields{
			"id":    ch.GetId(),
			"bytes": pkt.Nbytes,
			"seq":   pkt.Seq,
			"rtt":   pkt.Rtt,
		}).Debug("Received ping packet")
	})

	log.WithFields(log.Fields{
		"id":         ch.GetId(),
		"count":      pinger.Count(),
		"timeoutSec": ch.Timeout,
		"timeoutDur": pinger.Timeout(),
	}).Debug("Starting pinger")

	// blocking
	pinger.Run()

	stats := pinger.Statistics()

	cr := NewCheckResult(
		metric.NewPercentMetricFromInt("available", "", stats.PacketsRecv, stats.PacketsSent),
		metric.NewMetric("average", "", metric.MetricFloat, utils.ScaleFractionalDuration(stats.AvgRtt, time.Second), metric.UnitSeconds),
		metric.NewMetric("count", "", metric.MetricNumber, stats.PacketsSent, ""),
		metric.NewMetric("maximum", "", metric.MetricFloat, utils.ScaleFractionalDuration(stats.MaxRtt, time.Second), metric.UnitSeconds),
		metric.NewMetric("minimum", "", metric.MetricFloat, utils.ScaleFractionalDuration(stats.MinRtt, time.Second), metric.UnitSeconds),
	)
	crs := NewCheckResultSet(ch, cr)
	if stats.PacketsSent == 0 {
		log.Warn("No ping packets were sent, likely due to lack of permission")
		crs.SetStateUnavailable()
	} else {
		crs.SetStateAvailable()
	}

	log.WithFields(log.Fields{
		"id":     ch.GetId(),
		"stats":  stats,
		"result": cr,
	}).Debug("Finished remote.ping check")

	return crs, nil
}
