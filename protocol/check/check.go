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

// Package check provides the messaging structures specific to poller checks
package check

import (
	"encoding/json"
)

type CheckHeader struct {
	Id             string            `json:"id"`
	CheckType      string            `json:"type"`
	Period         uint64            `json:"period"`
	Timeout        uint64            `json:"timeout"`
	EntityId       string            `json:"entity_id"`
	ZoneId         string            `json:"zone_id"`
	Disabled       bool              `json:"disabled"`
	IpAddresses    map[string]string `json:"ip_addresses"`
	TargetAlias    *string           `json:"target_alias"`
	TargetHostname *string           `json:"target_hostname"`
	TargetResolver *string           `json:"target_resolver"`
}

// CheckIn is used for unmarshalling received check requests.
// Since the details are polymorphic, they are captured for delayed unmarshalling via the RawDetails field.
type CheckIn struct {
	CheckHeader

	RawDetails *json.RawMessage `json:"details"`
}
