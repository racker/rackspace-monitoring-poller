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


package endpoint

import (
	"github.com/racker/rackspace-monitoring-poller/config"
	"crypto/tls"
)

type EndpointServer interface {
	ApplyConfig(cfg *config.EndpointConfig) error

	ListenAndServe() error
}


func LoadCertificateFromConfig(cfg config.EndpointConfig) (*tls.Certificate, error) {
	if cfg.CertFile == "" {
		return nil, config.BadConfig{Details: "Missing CertFile"}
	}
	if cfg.KeyFile == "" {
		return nil, config.BadConfig{Details: "Missing KeyFile"}
	}

	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, err
	}

	return &cert, nil
}
