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

type HostInfoSystemMetrics struct {
	Arch          string `json:"arch"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	VendorName    string `json:"vendor_name"`
	Vendor        string `json:"vendor"`
	VendorVersion string `json:"vendor_version"`
}

type HostInfoSystemResult struct {
	Metrics   HostInfoSystemMetrics `json:"metrics"`
	Timestamp int64                 `json:"timestamp"`
}
