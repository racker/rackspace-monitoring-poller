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
	"context"
	"encoding/json"
	"log"
)

// NewCheck will unmarshal the request into one of the known polymorphic types given a received check request.
// This method needs to be updated to add to the known types.
func NewCheck(checkCtx context.Context, rawParams json.RawMessage, cancel context.CancelFunc) Check {
	checkBase := &Base{
		context: checkCtx,
		cancel:  cancel,
	}
	err := json.Unmarshal(rawParams, &checkBase)
	if err != nil {
		log.Printf("Error unmarshalling checkbase")
		return nil
	}
	switch checkBase.CheckType {
	case "remote.tcp":
		return NewTCPCheck(checkBase)
	case "remote.http":
		return NewHTTPCheck(checkBase)
	case "remote.ping":
		return NewPingCheck(checkBase)
	default:
		log.Printf("Invalid check type: %v", checkBase.CheckType)
	}
	return nil
}
