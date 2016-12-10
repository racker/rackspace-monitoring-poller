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

type HostInfoProcessesMetrics struct {
	Pid       int32   `json:"pid"`
	ExeName   string  `json:"exe_name"`
	ExeCwd    string  `json:"exe_cwd"`
	ExeRoot   string  `json:"exe_root"`
	StartTime int64   `json:"time_start_time"`
	TimeUser  float64 `json:"time_user"`
	TimeSys   float64 `json:"time_sys"`
	TimeTotal float64 `json:"time_total"`
	StateName string  `json:"state_name"`
	MemoryRes uint64  `json:"memory_resident"`
}

type HostInfoProcessesResult struct {
	Metrics   []HostInfoProcessesMetrics `json:"metrics"`
	Timestamp int64                      `json:"timestamp"`
}
