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
	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/racker/rackspace-monitoring-poller/config"
	"net"
	"net/url"
	"os"
	"time"
)

const (
	defaultPrometheusPushGatewayPort = "9091"
	prometheusService                = "prometheus"
	prometheusProto                  = "tcp"

	metricLabelCheckType = "check_type"
	metricLabelAddress   = "address"
	metricLabelZone      = "zone"
)

var (
	metricsRegistry = prometheus.NewRegistry()
)

func StartMetricsPusher(ctx context.Context, cfg *config.Config) {
	go runMetricsPusher(ctx, cfg)

}

func runMetricsPusher(ctx context.Context, cfg *config.Config) {
	log.Debug("Metrics pusher waiting to start...")
	defer log.Debug("Metric pusher exiting")

	time.Sleep(10 * time.Second)

	if cfg.PrometheusUri != "" {
		runPrometheusMetricsPusher(ctx, cfg)
	}
	// ...other metrics consumers can be added here
}

func runPrometheusMetricsPusher(ctx context.Context, cfg *config.Config) {

	gatewayUri, err := url.Parse(cfg.PrometheusUri)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"uri": cfg.PrometheusUri,
		}).Warn("Failed to parse Promtheus push gateway URI")
	}

	log.WithField("uri", gatewayUri).Info("Pushing metrics to Prometheus gateway")

	var promPushGateway string
	switch gatewayUri.Scheme {
	case "srv":
		_, addrs, err := net.LookupSRV(prometheusService, prometheusProto, gatewayUri.Hostname())
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
				"uri": gatewayUri,
			}).Warn("Failed to resolve Prometheus gateway service")
			return // TODO, retry at a later time
		}

		if len(addrs) == 0 {
			log.WithFields(log.Fields{
				"uri": gatewayUri,
			}).Warn("No addresses resolved for Prometheus gateway service")
			return // TODO, retry at a later time
		}

		promPushGateway = net.JoinHostPort(addrs[0].Target, string(addrs[0].Port))

	case "tcp":
		port := gatewayUri.Port()
		if port == "" {
			port = defaultPrometheusPushGatewayPort
		}
		promPushGateway = net.JoinHostPort(gatewayUri.Hostname(), port)

	default:
		log.WithField("uri", gatewayUri).Warn("Unsupported Prometheus gateway URI scheme")
		return
	}

	metricsRegistry.MustRegister(prometheus.NewGoCollector())

	groupings := map[string]string{
		"instance": cfg.Guid,
	}
	hostname, err := os.Hostname()
	if err == nil {
		groupings["hostname"] = hostname
	} else {
		log.WithField("err", err).Debug("Failed to get our hostname")
	}
	if cfg.AgentId != "" {
		groupings["agentId"] = cfg.AgentId
	}

	log.Debug("Metrics pusher started")
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			pushPrometheusMetrics(cfg, promPushGateway, groupings)
		}
	}
}

func pushPrometheusMetrics(cfg *config.Config, promPushGateway string, groupings map[string]string) {
	err := push.FromGatherer(cfg.AgentName, groupings, promPushGateway, metricsRegistry)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err,
			"gateway": promPushGateway,
		}).Warn("Failed to push metrics to Prometheus gateway")
	}
}
