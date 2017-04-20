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
	"errors"
	"fmt"
	"github.com/racker/rackspace-monitoring-poller/protocol/check"
)

// NewCheck will unmarshal the request into one of the known polymorphic types given a received check request.
// This method needs to be updated to add to the known types.
func NewCheck(parentContext context.Context, rawParams json.RawMessage) (Check, error) {
	ctx, cancel := context.WithCancel(parentContext)

	checkBase := &Base{
		context: ctx,
		cancel:  cancel,
	}
	err := json.Unmarshal(rawParams, &checkBase)
	if err != nil {
		return nil, err
	}

	return resolveCheckType(checkBase)
}

func NewCheckParsed(parentContext context.Context, checkIn check.CheckIn) (Check, error) {
	ctx, cancel := context.WithCancel(parentContext)

	checkBase := &Base{
		CheckIn: checkIn,
		context: ctx,
		cancel:  cancel,
	}

	return resolveCheckType(checkBase)
}

func resolveCheckType(checkBase *Base) (Check, error) {
	switch checkBase.CheckType {
	case "remote.tcp":
		return NewTCPCheck(checkBase)
	case "remote.http":
		return NewHTTPCheck(checkBase)
	case "remote.ping":
		return NewPingCheck(checkBase)
	}
	return nil, errors.New(fmt.Sprintf("Invalid check type: %v", checkBase.CheckType))
}
