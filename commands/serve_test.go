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

package commands_test

import (
	"github.com/racker/rackspace-monitoring-poller/commands"
	"github.com/racker/rackspace-monitoring-poller/config"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestServe_tls_defaults(t *testing.T) {
	assert := assert.New(t)

	certPool := commands.LoadRootCAs(false)
	if assert.NotNil(certPool) {
		assert.Len(certPool.Subjects(), 2)
	}
}

func TestServe_tls_insecure(t *testing.T) {
	assert := assert.New(t)

	certPool := commands.LoadRootCAs(true)
	assert.Nil(certPool)
}

func TestServe_tls_staging(t *testing.T) {
	assert := assert.New(t)

	os.Setenv(config.EnvStaging, config.EnabledEnvOpt)
	certPool := commands.LoadRootCAs(false)
	if assert.NotNil(certPool) {
		assert.Len(certPool.Subjects(), 1)
	}
}

func TestServe_tls_development(t *testing.T) {
	assert := assert.New(t)

	os.Setenv(config.EnvDevCA, "testdata/ca.pem")
	certPool := commands.LoadRootCAs(false)
	if assert.NotNil(certPool) {
		assert.Len(certPool.Subjects(), 1)
	}
}
