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
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
	protocol "github.com/racker/rackspace-monitoring-poller/protocol/check"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/sparrc/go-ping"
)

const (
	defaultPingCount = 6
)

// PingCheck conveys Ping checks
type PingCheck struct {
	Base
	protocol.PingCheckDetails
}

func NewPingCheck(base *Base) (Check, error) {
	check := &PingCheck{Base: *base}
	err := json.Unmarshal(*base.RawDetails, &check.Details)
	if err != nil {
		log.WithFields(log.Fields{
			"prefix":  "check_ping",
			"err":     err,
			"details": string(*base.RawDetails),
		}).Error("Unable to unmarshal check details")
		return nil, err
	}
	return check, nil
}

// Run method implements Check.Run method for Ping
// please see Check interface for more information
func (ch *PingCheck) Run() (*ResultSet, error) {
	log.Debugf("Running PING Check: %v", ch.GetID())

	targetIP, err := ch.GetTargetIP()
	if err != nil {
		return nil, err
	}

	var ipVersion string
	switch ch.TargetResolver {
	case protocol.ResolverIPV6:
		ipVersion = "v6"
	case protocol.ResolverIPV4:
		ipVersion = "v4"
	}

	pinger, err := ping.NewPinger(targetIP)
	if err != nil {
		log.WithFields(log.Fields{
			"prefix":   ch.GetLogPrefix(),
			"targetIP": targetIP,
		}).Error("Failed to create pinger")
		return nil, err
	}

	timeoutDuration := time.Duration(ch.Timeout) * time.Second

	count := int(ch.Details.Count)
	if count <= 0 {
		count = defaultPingCount
	}
	interPingDelay := utils.MinOfDurations(1*time.Second, timeoutDuration/time.Duration(count))
	perPingDuration := timeoutDuration / time.Duration(count)

	log.WithFields(log.Fields{
		"prefix":          ch.GetLogPrefix(),
		"targetIP":        targetIP,
		"ipVersion":       ipVersion,
		"interPingDelay":  interPingDelay,
		"perPingDuration": perPingDuration,
	}).Info("Running check")

	pinger.Count = count
	pinger.Timeout = timeoutDuration
	pinger.Interval = perPingDuration

	pinger.Run()

	stats := pinger.Statistics()

	var totalRTT time.Duration
	for _, rtt := range stats.Rtts {
		totalRTT += rtt
	}

	log.WithFields(log.Fields{
		"prefix":   ch.GetLogPrefix(),
		"checkId":  ch.Id,
		"targetIP": targetIP,
		"stats":    stats,
	}).Debug("Computed overall ping results")

	cr := NewResult(
		metric.NewMetric("count", "", metric.MetricNumber, stats.PacketsSent, ""),
		metric.NewPercentMetricFromInt("available", "", stats.PacketsRecv, stats.PacketsSent),
		metric.NewMetric("average", "", metric.MetricFloat, utils.ScaleFractionalDuration(stats.AvgRtt, time.Second), metric.UnitSeconds),
		metric.NewMetric("maximum", "", metric.MetricFloat, utils.ScaleFractionalDuration(stats.MaxRtt, time.Second), metric.UnitSeconds),
		metric.NewMetric("minimum", "", metric.MetricFloat, utils.ScaleFractionalDuration(stats.MinRtt, time.Second), metric.UnitSeconds),
	)
	crs := NewResultSet(ch, cr)

	if stats.PacketsRecv == 0 {
		crs.SetStateUnavailable()
		crs.SetStatus("No responses received")
	} else {
		crs.SetStateAvailable()
	}

	return crs, nil
}
