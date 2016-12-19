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

package config

import (
	"encoding/json"
	"io/ioutil"

	"os"

	log "github.com/Sirupsen/logrus"
)

type EndpointConfig struct {
	CertFile string
	KeyFile  string

	// In the form of "IP:port" or just ":port" to bind to all interfaces
	BindAddr string

	// StatsDAddr specifies the UDP host:port of a StatsD backend
	StatsDAddr string

	// AgentsConfigDir references a directory that is structured as
	//   <ZONE>/
	//     <AGENT ID>/
	//       checks/
	//         *.json
	AgentsConfigDir string
}

func NewEndpointConfig() *EndpointConfig {
	return &EndpointConfig{}
}

func (cfg *EndpointConfig) LoadFromFile(filepath string) error {
	configFile, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer configFile.Close()

	content, err := ioutil.ReadAll(configFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, cfg)
	if err != nil {
		return err
	}

	log.WithField("file", filepath).Info("Loaded configuration")

	return nil
}
