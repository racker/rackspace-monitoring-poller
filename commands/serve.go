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

package commands

import (
	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/poller"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
	"crypto/x509"
)

var (
	configFilePath    string
	insecure          bool

	ServeCmd          = &cobra.Command{
		Use:   "serve",
		Short: "Start the service",
		Long:  "Start the service",
		Run:   serveCmdRun,
	}
)

func init() {
	ServeCmd.Flags().StringVar(&configFilePath, "config", "", "Path to a file containing the config, used in "+config.DefaultConfigPathLinux)
	ServeCmd.Flags().BoolVar(&insecure, "insecure", false, "Enabled ONLY during development and when connecting to a known farend")
}

func HandleInterrupts() chan os.Signal {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	return c
}

func LoadRootCAs(insecure bool) *x509.CertPool {
	if insecure {
		log.Warn("Insecure TLS connectivity is enabled. Make sure you TRUST the farend host")
		return nil
	} else {
		devCAPath := os.Getenv(config.EnvDevCA)
		isStaging := os.Getenv(config.EnvStaging)
		if isStaging == "1" {
			log.Warn("Staging root CAs are in use")
			return config.LoadStagingCAs()
		} else if devCAPath != "" {
			log.WithField("path", devCAPath).Warn("Development root CAs are in use")
			return config.LoadDevelopmentCAs(devCAPath)
		} else {
			return config.LoadProductionCAs()
		}
	}
}

func serveCmdRun(cmd *cobra.Command, args []string) {
	guid := uuid.NewV4()
	cfg := config.NewConfig(guid.String())
	err := cfg.LoadFromFile(configFilePath)
	if err != nil {
		utils.Die(err, "Failed to load configuration")
	}

	rootCAs := LoadRootCAs(insecure)

	log.WithField("guid", guid).Info("Assigned unique identifier")

	signalNotify := HandleInterrupts()
	for {
		stream := poller.NewConnectionStream(cfg, rootCAs)
		stream.Connect()
		waitCh := stream.WaitCh()
		for {
			select {
			case <-waitCh:
				break
			case <-signalNotify:
				log.Info("Shutdown...")
				stream.Stop()
			case <-stream.StopNotify():
				return
			}
		}
	}
}
