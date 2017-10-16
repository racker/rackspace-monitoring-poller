//
// Copyright 2017 Rackspace
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

package poller

import (
	"github.com/racker/rackspace-monitoring-poller/config"
	"context"
	"github.com/DataDog/datadog-go/statsd"
	log "github.com/sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/check"
	"fmt"
	"strings"
	"os"
)

type MetricsDistributor struct {
	types []metricsDistributorType
	ctx   context.Context
}

type metricsDistributorType interface {
	Start(ctx context.Context) error
	Distribute(crs *check.ResultSet)
}

func NewMetricsDistributor(ctx context.Context, config *config.Config) *MetricsDistributor {
	distributor := &MetricsDistributor{ctx: ctx}

	var tenantId string
	tokenParts := strings.Split(config.Token, ".")
	if len(tokenParts) >= 2 {
		tenantId = tokenParts[1]
	}

	if config.StatsdEndpoint != "" {
		statsdDistributor := newStatsdDistributor(config.StatsdEndpoint, tenantId)
		if statsdDistributor != nil {
			distributor.types = append(distributor.types, statsdDistributor)
		}
	}

	return distributor
}

func (md *MetricsDistributor) Start() {
	for _, d := range md.types {
		err := d.Start(md.ctx)
		if err != nil {
			log.WithError(err).WithField("type", d).Warn("Failed to start distributor type")
		}
	}
}

func (md *MetricsDistributor) Distribute(crs *check.ResultSet) {
	for _, d := range md.types {
		d.Distribute(crs)
	}
}

type statsdDistributor struct {
	statsdEndpoint string
	crsChan        chan *check.ResultSet
	statsdClient   *statsd.Client
	tenantId       string
	ctx            context.Context
}

func newStatsdDistributor(statsdEndpoint string, tenantId string) metricsDistributorType {
	d := &statsdDistributor{
		tenantId:       tenantId,
		statsdEndpoint: statsdEndpoint,
	}

	return d
}

func (d *statsdDistributor) String() string {
	return fmt.Sprintf("statsdDistributor[endpoint=%s]", d.statsdEndpoint)
}

func (d *statsdDistributor) Start(ctx context.Context) error {
	var err error
	d.statsdClient, err = statsd.New(d.statsdEndpoint)
	if err != nil {
		return err
	}
	d.statsdClient.Namespace = "rackspace."
	d.crsChan = make(chan *check.ResultSet, 100)

	log.WithField("endpoint", d.statsdEndpoint).Info("Using statsd metrics distribution type")

	go d.run(ctx)

	return nil
}

func (d *statsdDistributor) Distribute(crs *check.ResultSet) {
	d.crsChan <- crs
}

func (d *statsdDistributor) run(ctx context.Context) {
	for {
		select {
		case crs := <-d.crsChan:
			d.sendToStatsd(crs)

		case <-ctx.Done():
			return
		}
	}
}

func (d *statsdDistributor) sendToStatsd(crs *check.ResultSet) {

	ch := crs.Check
	checkType := ch.GetCheckType()
	checkID := ch.GetID()
	entityId := ch.GetEntityID()
	state := crs.State
	tags := []string{
		"tenant:" + d.tenantId,
		"entity:" + entityId,
		"check:" + checkID,
	}

	if targetIp, err := ch.GetTargetIP(); err != nil {
		tags = append(tags, "targetIP:"+targetIp)
	}

	hostname, hostnameErr := os.Hostname()
	if hostnameErr != nil {
		log.WithError(hostnameErr).Warn("Unable to identify our own hostname")
	} else {
		tags = append(tags, "host:"+hostname)
	}

	for _, result := range crs.Metrics {
		for name, m := range result.Metrics {
			metricName := fmt.Sprintf("%s.%s", checkType, name)
			value, err := m.ToFloat64()
			if err == nil {
				log.WithField("metric", metricName).Debug("Sending metric to statsd")
				d.statsdClient.Gauge(metricName, value, tags, 1)
			} else {
				log.WithField("metric", metricName).Debug("Failed to derive float64 value")
			}
		}
	}

	serviceName := checkType
	serviceCheck := statsd.NewServiceCheck(serviceName, mapStateToServiceCheckStatus(state))
	serviceCheck.Message = crs.Status
	serviceCheck.Tags = tags
	log.WithField("service", serviceName).Debug("Sending service check to statsd")
	d.statsdClient.ServiceCheck(serviceCheck)
}

func mapStateToServiceCheckStatus(state string) statsd.ServiceCheckStatus {
	switch state {
	case check.StateAvailable:
		return statsd.Ok
	case check.StateUnavailable:
		return statsd.Critical
	}

	return statsd.Unknown
}
