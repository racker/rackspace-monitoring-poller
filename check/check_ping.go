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

	log "github.com/Sirupsen/logrus"
	protocol "github.com/racker/rackspace-monitoring-poller/protocol/check"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"
)

const (
	defaultPingCount = 5
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

	pinger, err := PingerFactory(ch.GetID(), targetIP, ipVersion)
	if err != nil {
		log.WithFields(log.Fields{
			"prefix":   ch.GetLogPrefix(),
			"targetIP": targetIP,
		}).Error("Failed to create pinger")
		return nil, err
	}
	defer pinger.Close()

	timeoutDuration := time.Duration(ch.Timeout) * time.Second
	overallTimeout := time.After(timeoutDuration)

	count := int(ch.Details.Count)
	if count <= 0 {
		count = defaultPingCount
	}
	interPingDelay := utils.MinOfDurations(1*time.Second, timeoutDuration/time.Duration(count))
	perPingDuration := time.Duration((ch.Timeout*1000)/uint64(count)) * time.Millisecond

	rtts := make([]time.Duration, 0)
	var pingErr error

packetLoop:
	for i := 0; i < count; i++ {
		seq := i + 1

		select {
		case resp := <-pinger.Ping(seq):
			log.WithFields(log.Fields{
				"prefix":  ch.GetLogPrefix(),
				"resp":    resp,
				"checkId": ch.Id,
			}).Debug("Got ping response")

			if resp.Err != nil {
				pingErr = resp.Err
				break packetLoop
			}
			if resp.Seq != seq {
				log.WithFields(log.Fields{
					"prefix": ch.GetLogPrefix(),
					"seq":    resp.Seq,
				}).Debug("Incorrect packet seq received")
			} else {
				rtts = append(rtts, resp.Rtt)
			}
			time.Sleep(interPingDelay)

		case <-time.After(perPingDuration):
			log.WithFields(log.Fields{
				"prefix":   ch.GetLogPrefix(),
				"targetIP": targetIP,
				"seq":      seq,
			}).Debug("Timed out getting response")

		case <-overallTimeout:
			log.WithFields(log.Fields{
				"prefix":   ch.GetLogPrefix(),
				"targetIP": targetIP,
			}).Debug("Reached overall timeout")

			break packetLoop
		}

	}

	sent := count
	recv := len(rtts)
	var minRTT time.Duration
	var maxRTT time.Duration
	var avgRTT time.Duration

	if recv > 0 {
		minRTT = rtts[0]
		maxRTT = rtts[0]
		var totalRTT time.Duration
		for _, rtt := range rtts {
			totalRTT += rtt
			if rtt > maxRTT {
				maxRTT = rtt
			}
			if rtt < minRTT {
				minRTT = rtt
			}
		}
		avgRTT = totalRTT / time.Duration(recv)
	}
	log.WithFields(log.Fields{
		"prefix":   ch.GetLogPrefix(),
		"checkId":  ch.Id,
		"targetIP": targetIP,
		"sent":     sent,
		"recv":     recv,
		"minRTT":   minRTT,
		"maxRTT":   maxRTT,
		"avgRTT":   avgRTT,
	}).Debug("Computed overall ping results")

	cr := NewResult(
		metric.NewMetric("count", "", metric.MetricNumber, sent, ""),
		metric.NewPercentMetricFromInt("available", "", recv, sent),
		metric.NewMetric("average", "", metric.MetricFloat, utils.ScaleFractionalDuration(avgRTT, time.Second), metric.UnitSeconds),
		metric.NewMetric("maximum", "", metric.MetricFloat, utils.ScaleFractionalDuration(maxRTT, time.Second), metric.UnitSeconds),
		metric.NewMetric("minimum", "", metric.MetricFloat, utils.ScaleFractionalDuration(minRTT, time.Second), metric.UnitSeconds),
	)
	crs := NewResultSet(ch, cr)

	if pingErr == nil && recv > 0 {
		crs.SetStatusSuccess()
		crs.SetStateAvailable()
	} else if pingErr == nil && recv == 0 {
		crs.SetStateUnavailable()
		crs.SetStatus("No responses received")
	} else {
		crs.SetStateUnavailable()
		crs.SetStatusFromError(pingErr)
	}

	return crs, nil
}
