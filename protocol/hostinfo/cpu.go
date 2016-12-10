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

type HostInfoCpuMetrics struct {
	Idle         uint64  `json:"idle"`
	Irq          float64 `json:"irq"`
	Mhz          float64 `json:"mhz"`
	Model        string  `json:"model"`
	Name         string  `json:"name"`
	Nice         float64 `json:"nice"`
	SoftIrq      float64 `json:"softirq"`
	Stolen       float64 `json:"stolen"`
	Sys          float64 `json:"sys"`
	Total        uint64  `json:"total"`
	TotalCores   uint64  `json:"total_cores"`
	TotalSockets uint64  `json:"total_sockets"`
	User         float64 `json:"user"`
	Vendor       string  `json:"vendor"`
	Wait         float64 `json:"wait"`
}

type HostInfoCpuResult struct {
	Metrics   []HostInfoCpuMetrics `json:"metrics"`
	Timestamp int64                `json:"timestamp"`
}
