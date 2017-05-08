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
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/satori/go.uuid"
	"os"
	"time"
)

func generatePollerGuid() string {
	return uuid.NewV4().String()
}

func Run(configFilePath string, insecure bool) {
	utils.CheckFDLimit()

	guid := generatePollerGuid()
	useStaging := config.IsUsingStaging()
	features := []config.Feature{
		{Name: "poller", Disabled: false},
	}

	cfg := config.NewConfig(guid, useStaging, features)
	if err := cfg.LoadFromFile(configFilePath); err != nil {
		utils.Die(err, "Failed to load configuration")
	}
	if err := cfg.Validate(); err != nil {
		utils.Die(err, "Failed to validate configuration")
	}

	log.WithField("guid", guid).Info("Assigned unique identifier")

	rootCAs := config.LoadRootCAs(insecure, useStaging)

	signalNotify := utils.HandleInterrupts()
	ctx, cancel := context.WithCancel(context.Background())
	StartMetricsPusher(ctx, cfg)
	for {
		stream := NewConnectionStream(ctx, cfg, rootCAs)
		stream.Connect()
		for {
			select {
			case <-signalNotify:
				log.Info("Shutdown...")
				cancel()
				time.AfterFunc(gracefulShutdownTimeout, func() {
					log.Warn("Forcing immediate shutdown")
					os.Exit(0)
				})
			case <-stream.Done(): // for cancel to propagate
				return
			}
		}
	}

}
