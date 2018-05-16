/*
 *
 * Copyright 2018 Rackspace
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS-IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package commands

import (
	"context"

	"github.com/kardianos/service"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/poller"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	configFilePath string
	insecure       bool

	svcConfig = &service.Config{
		Name:        "Rackspace Monitoring Poller",
		DisplayName: "Rackspace Monitoring Poller",
		Description: "Provides availablity monitoring for devices on a private network",
	}

	ServeCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the service",
		Long:  "Start the service",
		Run:   serveCmdRun,
	}

	serviceContext context.Context
	serviceCancel  context.CancelFunc
)

func init() {
	ServeCmd.Flags().StringVar(&configFilePath, "config", config.DefaultConfigPathLinux, "Path to a file containing the config")
	ServeCmd.Flags().BoolVar(&insecure, "insecure", false, "Enabled ONLY during development and when connecting to a known farend")
}

func serveCmdRun(cmd *cobra.Command, args []string) {
	if service.Interactive() {
		poller.Run(nil, configFilePath, insecure)
	} else {
		s, logger, err := createServiceInstance()
		if err != nil {
			log.Fatal(err)
		}

		err = s.Run()
		if err != nil {
			logger.Error(err)
		}
	}

}

func createServiceInstance() (service.Service, service.Logger, error) {
	entry := &serviceEntry{}
	s, err := service.New(entry, svcConfig)
	if err != nil {
		return nil, nil, err
	}

	logger, err := s.Logger(nil)
	if err != nil {
		return nil, nil, err
	}

	return s, logger, nil
}

type serviceEntry struct{}

func (se *serviceEntry) Start(s service.Service) error {
	serviceContext, serviceCancel = context.WithCancel(context.Background())
	go poller.Run(serviceContext, configFilePath, insecure)

	return nil
}

func (se *serviceEntry) Stop(s service.Service) error {
	serviceCancel()
	return nil
}
