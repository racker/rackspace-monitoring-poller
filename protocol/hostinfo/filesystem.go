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

type HostInfoFilesystemMetrics struct {
	DirectoryName  string `json:"dir_name"`
	DeviceName     string `json:"dev_name"`
	Options        string `json:"options"`
	SystemTypeName string `json:"sys_type_name"`
	Free           uint64 `json:"free"`
	Total          uint64 `json:"total"`
	Used           uint64 `json:"used"`
	Available      uint64 `json:"avail"`
	Files          uint64 `json:"files"`
	FreeFiles      uint64 `json:"free_files"`
}

type HostInfoFilesystemResult struct {
	Metrics   []HostInfoFilesystemMetrics `json:"metrics"`
	Timestamp int64                       `json:"timestamp"`
}
