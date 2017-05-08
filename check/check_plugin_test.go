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
	  "target_resolver":"IPv4",
	  "disabled":false
	  }`, "fixtures/cloudkick_agent_custom_plugin_1.sh")
	check, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	crs, err := check.Run()
	require.NoError(t, err)
	require.True(t, crs.Available, "validate errored")

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
	  "target_resolver":"IPv4",
	  "disabled":false
	  }`, "fixtures/cloudkick_agent_custom_plugin_timeout.sh")
	check, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	crs, err := check.Run()
	require.NoError(t, err)

	require.False(t, crs.Available, "validate errored")
}

func TestAgentPlugin_WarnPlugin2(t *testing.T) {
	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chTestPlugin",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"file":"%s"},
	  "type":"agent.plugin",
	  "timeout":5,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"IPv4",
	  "disabled":false
	  }`, "fixtures/cloudkick_agent_custom_plugin_2.sh")
	check, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	crs, err := check.Run()
	require.NoError(t, err)

	require.True(t, crs.Available, "status is warn")
}

func TestAgentPlugin_InvalidMetricLines1(t *testing.T) {
	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chTestPlugin",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"file":"%s"},
	  "type":"agent.plugin",
	  "timeout":5,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"IPv4",
	  "disabled":false
	  }`, "fixtures/invalid_metric_lines_1.sh")
	check, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	crs, err := check.Run()
	require.NoError(t, err)

	require.True(t, crs.Available, "status is invalid")
	cr := crs.Get(0)
	require.Nil(t, cr.GetMetric("metric1"), "validate metric1 metric")
	require.Nil(t, cr.GetMetric("foo"), "validate foo metric")
	require.NotNil(t, cr.GetMetric("metric8"), "validate metric8 metric")
	require.NotNil(t, cr.GetMetric("metric7"), "validate metric7 metric")
	require.NotNil(t, cr.GetMetric("metric2"), "validate metric2 metric")
}

func TestAgentPlugin_InvalidMetricLines2(t *testing.T) {
	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chTestPlugin",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"file":"%s"},
	  "type":"agent.plugin",
	  "timeout":5,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"IPv4",
	  "disabled":false
	  }`, "fixtures/invalid_metric_lines_2.sh")
	check, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	crs, err := check.Run()
	require.NoError(t, err)

	require.True(t, crs.Available, "status is invalid")
	cr := crs.Get(0)
	require.Nil(t, cr.GetMetric("foo"), "validate foo metric")
	require.Nil(t, cr.GetMetric("metric1"), "validate metric1 metric")
	require.NotNil(t, cr.GetMetric("metric7"), "validate metric7 metric")
	require.NotNil(t, cr.GetMetric("metric8"), "validate metric8 metric")
	require.NotNil(t, cr.GetMetric("metric2"), "validate metric1 metric")
	require.NotNil(t, cr.GetMetric("metric2"), "validate metric2 metric")
}

func TestAgentPlugin_InvalidMetricLines3(t *testing.T) {
	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chTestPlugin",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"file":"%s"},
	  "type":"agent.plugin",
	  "timeout":5,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"IPv4",
	  "disabled":false
	  }`, "fixtures/invalid_metric_lines_3.sh")
	check, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	crs, err := check.Run()
	require.NoError(t, err)

	require.True(t, crs.Available, "status is invalid")
	cr := crs.Get(0)
	require.Nil(t, cr.GetMetric("foo"), "validate foo metric")
	require.Nil(t, cr.GetMetric("metric1"), "validate metric1 metric")
	require.NotNil(t, cr.GetMetric("metric8"), "validate metric8 metric")
	require.NotNil(t, cr.GetMetric("metric7"), "validate metric7 metric")
	require.NotNil(t, cr.GetMetric("metric2"), "validate metric1 metric")
	require.NotNil(t, cr.GetMetric("metric2"), "validate metric2 metric")
}

func TestAgentPlugin_InvalidMetricLines4(t *testing.T) {
	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chTestPlugin",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"file":"%s"},
	  "type":"agent.plugin",
	  "timeout":5,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"IPv4",
	  "disabled":false
	  }`, "fixtures/invalid_metric_lines_4.sh")
	check, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	crs, err := check.Run()
	require.NoError(t, err)

	require.True(t, crs.Available, "status is invalid")
	cr := crs.Get(0)
	require.Nil(t, cr.GetMetric("foo"), "validate foo metric")
	require.Nil(t, cr.GetMetric("metric1"), "validate metric1 metric")
	require.NotNil(t, cr.GetMetric("metric8"), "validate metric8 metric")
	require.NotNil(t, cr.GetMetric("metric7"), "validate metric7 metric")
	require.NotNil(t, cr.GetMetric("metric2"), "validate metric1 metric")
	require.NotNil(t, cr.GetMetric("metric2"), "validate metric2 metric")
}

func TestAgentPlugin_StatusCodeNonZero(t *testing.T) {
	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chTestPlugin",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"file":"%s"},
	  "type":"agent.plugin",
	  "timeout":5,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"IPv4",
	  "disabled":false
	  }`, "fixtures/non_zero_with_status.sh")
	ch, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	crs, err := ch.Run()
	require.NoError(t, err)

	require.False(t, crs.Available, "status is invalid")
	require.Equal(t, crs.Status, check.ErrorPluginExit)
}

func TestAgentPlugin_NotExecutable(t *testing.T) {
	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chTestPlugin",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"file":"%s"},
	  "type":"agent.plugin",
	  "timeout":5,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"IPv4",
	  "disabled":false
	  }`, "fixtures/not_executable.sh")
	ch, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	crs, err := ch.Run()
	require.NoError(t, err)

	require.False(t, crs.Available, "status is invalid")
	require.Equal(t, crs.Status, check.ErrorPluginExit)
}

func TestAgentPlugin_PartialOutputWithSleep(t *testing.T) {
	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chTestPlugin",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"file":"%s"},
	  "type":"agent.plugin",
	  "timeout":5,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"IPv4",
	  "disabled":false
	  }`, "fixtures/partial_output_with_sleep.sh")
	ch, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	crs, err := ch.Run()
	require.NoError(t, err)

	cr := crs.Get(0)
	require.True(t, crs.Available)
	require.NotNil(t, cr.GetMetric("logged_users"))
	require.NotNil(t, cr.GetMetric("active_processes"))
	require.NotNil(t, cr.GetMetric("avg_wait_time"))
	require.NotNil(t, cr.GetMetric("something"))
	require.NotNil(t, cr.GetMetric("packet_count"))
}

func TestAgentPlugin_Plugin1(t *testing.T) {
	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chTestPlugin",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"file":"%s"},
	  "type":"agent.plugin",
	  "timeout":5,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"IPv4",
	  "disabled":false
	  }`, "fixtures/plugin_1.sh")
	ch, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	crs, err := ch.Run()
	require.NoError(t, err)

	cr := crs.Get(0)
	require.True(t, crs.Available)
	require.NotNil(t, cr.GetMetric("logged_users"))
	require.NotNil(t, cr.GetMetric("active_processes"))
	require.NotNil(t, cr.GetMetric("avg_wait_time"))
	require.NotNil(t, cr.GetMetric("something"))
	require.NotNil(t, cr.GetMetric("packet_count"))
}

func TestAgentPlugin_PluginCustomArguments(t *testing.T) {
	// Create Check
	checkData := fmt.Sprintf(`{
	  "id":"chTestPlugin",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
		"details":{"file":"%s", "args": ["a", "b"]},
	  "type":"agent.plugin",
	  "timeout":5,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"IPv4",
	  "disabled":false
	  }`, "fixtures/plugin_custom_arguments.sh")
	ch, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	crs, err := ch.Run()
	require.NoError(t, err)

	cr := crs.Get(0)
	require.True(t, crs.Available)
	require.NotNil(t, cr.GetMetric("a"))
	require.NotNil(t, cr.GetMetric("b"))
	require.Nil(t, cr.GetMetric("c"))
}
