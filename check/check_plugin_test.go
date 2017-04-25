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
package check_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/stretchr/testify/require"
)

func TestAgentPlugin_RunSuccess(t *testing.T) {
	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chTestPlugin",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"file":"%s"},
	  "type":"agent.plugin",
	  "timeout":15,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":1,
	  "disabled":false
	  }`, "fixtures/cloudkick_agent_custom_plugin_1.sh")
	check, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	// Run check
	crs, err := check.Run()
	require.NoError(t, err)
	require.False(t, crs.Available, "validate errored")

	cr := crs.Get(0)
	require.NotNil(t, cr.GetMetric("logged_users"), "validate logged_users metric")
	require.NotNil(t, cr.GetMetric("active_processes"), "validate logged_users metric")
}

func TestAgentPlugin_RunTimeout(t *testing.T) {
	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chTestPlugin",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"file":"%s"},
	  "type":"agent.plugin",
	  "timeout":1,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":1,
	  "disabled":false
	  }`, "fixtures/cloudkick_agent_custom_plugin_timeout.sh")
	check, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	// Run check
	crs, err := check.Run()
	require.NoError(t, err)

	require.False(t, crs.Available, "validate errored")
}
