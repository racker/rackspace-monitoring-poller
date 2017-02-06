//
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
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewMetricsPostRequest(t *testing.T) {
	checkData := `{
	  "id":"chTestTCP_TLSRunSuccess",
	  "zone_id":"pzA",
	  "entity_id":"enAAAAIPV4",
	  "details":{"port":443,"ssl":true},
	  "type":"remote.tcp",
	  "timeout":15,
	  "period":30,
	  "ip_addresses":{"default":"127.0.0.1"},
	  "target_alias":"default",
	  "target_hostname":"",
	  "target_resolver":"",
	  "disabled":false
	  }`
	ch, err := check.NewCheck(context.Background(), []byte(checkData))
	require.NoError(t, err)

	cr := check.NewResult(metric.NewMetric("tt_connect", "", metric.MetricNumber, 500, metric.UnitMilliseconds))
	crs := check.NewResultSet(ch, cr)

	origTimeStamper := utils.InstallAlternateTimestampFunc(func() int64 {
		return 5000
	})
	defer utils.InstallAlternateTimestampFunc(origTimeStamper)

	req := check.NewMetricsPostRequest(crs, 200)

	assert.Equal(t, int64(5200), req.Params.Timestamp)
}
