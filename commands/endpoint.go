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
	"github.com/spf13/cobra"
	"github.com/racker/rackspace-monitoring-poller/endpoint"
)

var (
	endpointConfigFilePath string
	EndpointCmd       = &cobra.Command{
		Use:   "endpoint",
		Short: "Start the endpoint service",
		Long:  "Start the endpoint service",
		Run:   endpointCmdRun,
	}
)

func init() {
	EndpointCmd.Flags().StringVar(&endpointConfigFilePath, "config", "", "Path to a file containing the endpoint config")
}

func endpointCmdRun(cmd *cobra.Command, args []string) {
	s, err := endpoint.NewEndpointServer(endpointConfigFilePath)

	if err != nil {
		log.Errorln("Invalid endpoint setup", err)
		return
	}

	err = s.ListenAndServe()
	if err != nil {
		log.Errorln("Endpoint serving failed", err)
	}
}
