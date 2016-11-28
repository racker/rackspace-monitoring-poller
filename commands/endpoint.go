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

// serve
package commands

import (
	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/types"
	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

var (
	configFilePath string
	EndpointCmd       = &cobra.Command{
		Use:   "endpoint",
		Short: "Start the endpoint service",
		Long:  "Start the endpoint service",
		Run:   endpointCmdRun,
	}
)

func init() {
	ServeCmd.Flags().StringVar(&configFilePath, "config", "", "Path to a file containing the config, used in "+config.DefaultConfigPathLinux)
}

func endpointCmdRun(cmd *cobra.Command, args []string) {
	guid := uuid.NewV4()
	log.Infof("Using GUID: %v", guid)
	cfg := config.NewConfig(guid.String())
	cfg.LoadFromFile(configFilePath)
	for {
		stream := types.NewConnectionStream(cfg)
		stream.Connect()
		stream.Wait()
	}
}
