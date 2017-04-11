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
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/racker/rackspace-monitoring-poller/poller"
	"github.com/spf13/cobra"
)

var (
	configFilePath string
	insecure       bool

	ServeCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the service",
		Long:  "Start the service",
		Run:   serveCmdRun,
	}
)

func init() {
	ServeCmd.Flags().StringVar(&configFilePath, "config", config.DefaultConfigPathLinux, "Path to a file containing the config")
	ServeCmd.Flags().BoolVar(&insecure, "insecure", false, "Enabled ONLY during development and when connecting to a known farend")
}

func serveCmdRun(cmd *cobra.Command, args []string) {
	poller.Run(configFilePath, insecure)
}
