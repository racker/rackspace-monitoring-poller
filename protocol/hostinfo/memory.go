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

package hostinfo

type HostInfoMemoryMetrics struct {
	UsedPercentage     float64 `json:"used_percentage"`
	Free               uint64  `json:"free"`
	Total              uint64  `json:"total"`
	Used               uint64  `json:"used"`
	SwapFree           uint64  `json:"swap_free"`
	SwapTotal          uint64  `json:"swap_total"`
	SwapUsed           uint64  `json:"swap_used"`
	SwapUsedPercentage float64 `json:"swap_percentage"`
}

type HostInfoMemoryResult struct {
	Metrics   HostInfoMemoryMetrics `json:"metrics"`
	Timestamp int64                 `json:"timestamp"`
}
