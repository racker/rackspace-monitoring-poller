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

package poller

import (
	"errors"
	"fmt"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"github.com/racker/rackspace-monitoring-poller/protocol/check"
)

// ActionableCheck enriches a check.CheckIn with an action indicator.
// It is expected that during preparation these may have only abbreviated
// fields populated in check.CheckIn.
type ActionableCheck struct {
	check.CheckIn

	Action string
	// Populated indicates if the check.CheckIn has been fully populated; however,
	// this is not applicable if Action is protocol.PrepareActionContinue.
	Populated bool
}

type CheckPreparation struct {
	TrackingVersion int

	// Actions is a 2D map of entityId->checkId->ActionableCheck
	Actions map[string] /*entityId*/ map[string] /*checkId*/ ActionableCheck
}

func NewCheckPreparation(version int, manifest []protocol.PollerPrepareManifest) *CheckPreparation {
	cp := &CheckPreparation{
		TrackingVersion: version,
		Actions:         make(map[string] /*entityId*/ map[string] /*checkId*/ ActionableCheck),
	}

	for _, m := range manifest {
		byEntity := cp.Actions[m.EntityId]
		if byEntity == nil {
			byEntity = make(map[string] /*checkId*/ ActionableCheck)
			cp.Actions[m.EntityId] = byEntity
		}

		byEntity[m.Id] = ActionableCheck{
			Action: m.Action,
			// "pre populate" actions we don't expect to see defined
			Populated: !checkPreparationNeedsPopulating(m.Action),

			CheckIn: check.CheckIn{
				CheckHeader: check.CheckHeader{
					ZoneId:    m.ZoneId,
					CheckType: m.CheckType,
				},
			},
		}
	}

	return cp
}

func checkPreparationNeedsPopulating(action string) bool {
	return action != protocol.PrepareActionContinue
}

func (cp *CheckPreparation) VersionApplies(version int) bool {
	return cp != nil && cp.TrackingVersion == version
}

func (cp *CheckPreparation) AddDefinitions(block []check.CheckIn) {

	for _, check := range block {
		actionable := cp.Actions[check.EntityId][check.Id]
		actionable.Populated = true
		actionable.CheckIn = check

		cp.Actions[check.EntityId][check.Id] = actionable
	}
}

func (cp *CheckPreparation) Validate(version int) error {
	if !cp.VersionApplies(version) {
		return errors.New("Wrong version")
	}

	for entityId, byEntity := range cp.Actions {
		for checkId, actionable := range byEntity {
			if actionable.Action == "" {
				return errors.New(fmt.Sprintf(
					"Check defined but not declared in manifest. entity=%v, check=%v",
					actionable.EntityId, actionable.Id))
			}

			if !actionable.Populated {
				return errors.New(fmt.Sprintf(
					"Check declared in manifest but not defined. entity=%v, check=%v",
					entityId, checkId))
			}
		}
	}

	return nil
}
