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

package check

import (
	"fmt"

	"github.com/racker/rackspace-monitoring-poller/protocol"
)

// NewMetricsPostRequest function sets up a request with provided
// check results set data. The timestamp within the request will be adjusted for the far-end's clock
// offset given as clockOffset, in milliseconds.
func NewMetricsPostRequest(crs *ResultSet, clockOffset int64) *protocol.MetricsPostRequest {
	req := &protocol.MetricsPostRequest{}
	req.Version = "1"
	req.Method = "check_metrics.post_multi"
	crs.PopulateMetricsPostContent(clockOffset, &req.Params)
	return req
}

// ConvertToMetricResults function iterates through the check result
// in check result set and format it to MetricTVU, add it to the list
// and return that list
func ConvertToMetricResults(cr *Result) protocol.MetricWrap {
	wrapper := make(map[string]*protocol.MetricTVU)
	for key, m := range cr.Metrics {
		wrapper[key] = &protocol.MetricTVU{
			Type:  m.TypeString,
			Value: fmt.Sprintf("%v", m.Value),
			Unit:  m.Unit,
		}
	}
	wrappers := protocol.MetricWrap{
		nil, // resource/dimension -- usused
		wrapper,
	}
	return wrappers
}
