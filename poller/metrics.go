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
	"time"
)

const (
	promPushGateway = "localhost:9091"
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

	log.Debug("Metrics pusher started")
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			pushMetrics(cfg)
		}
	}
}

func pushMetrics(cfg *config.Config) {
	push.FromGatherer(cfg.Guid, nil, promPushGateway, metricsRegistry)
}
