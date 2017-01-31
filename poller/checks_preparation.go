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
	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"github.com/racker/rackspace-monitoring-poller/protocol/check"
)

// ActionType provides a non-protocol enumeration of valid check preparation actions
type ActionType int

const (
	ActionTypeUnknown  = iota
	ActionTypeStart    = iota
	ActionTypeRestart  = iota
	ActionTypeContinue = iota
)

// ActionableCheck enriches a check.CheckIn with an action indicator.
// It is expected that during preparation these may have only abbreviated
// fields populated in check.CheckIn.
type ActionableCheck struct {
	check.CheckIn

	Action ActionType
	// Populated indicates if the check.CheckIn has been fully populated; however,
	// this is not applicable if Action is protocol.PrepareActionContinue.
	Populated bool
}

type ChecksPreparation struct {
	TrackingVersion int

	// Actions is a map of checkId->ActionableCheck
	actions map[string] /*checkId*/ ActionableCheck
}

// NewChecksPreparation initiates a new checks preparation session.
// Returns the new ChecksPreparation if successful or an error if an unsupported action type was encountered.
func NewChecksPreparation(version int, manifest []protocol.PollerPrepareManifest) (*ChecksPreparation, error) {
	cp := &ChecksPreparation{
		TrackingVersion: version,
		actions:         make(map[string]ActionableCheck),
	}

	for _, m := range manifest {

		actionType := mapToActionType(m.Action)
		if actionType == ActionTypeUnknown {
			return nil, errors.New(fmt.Sprintf("Unsupported action in manifest: action=%s, check=%s", m.Action, m.Id))
		}
		cp.actions[m.Id] = ActionableCheck{
			Action: actionType,
			// "pre populate" actions we don't expect to see defined
			Populated: !checkPreparationNeedsPopulating(m.Action),

			CheckIn: check.CheckIn{
				CheckHeader: check.CheckHeader{
					Id:        m.Id,
					EntityId:  m.EntityId,
					ZoneId:    m.ZoneId,
					CheckType: m.CheckType,
				},
			},
		}
	}

	log.WithField("actions", cp.actions).Debug("Prepared checks from manifest")
	return cp, nil
}

func mapToActionType(actionStr string) ActionType {
	switch actionStr {
	case protocol.PrepareActionStart:
		return ActionTypeStart
	case protocol.PrepareActionRestart:
		return ActionTypeRestart
	case protocol.PrepareActionContinue:
		return ActionTypeContinue
	default:
		return ActionTypeUnknown
	}
}

func checkPreparationNeedsPopulating(action string) bool {
	return action != protocol.PrepareActionContinue
}

func (cp *ChecksPreparation) GetActionableChecks() (actionableChecks []ActionableCheck) {
	actionableChecks = make([]ActionableCheck, 0, len(cp.actions))

	for _, ac := range cp.actions {
		actionableChecks = append(actionableChecks, ac)
	}

	return
}

func (cp *ChecksPreparation) VersionApplies(version int) bool {
	return cp != nil && cp.TrackingVersion == version
}

func (cp *ChecksPreparation) AddDefinitions(block []check.CheckIn) {

	for _, ch := range block {
		actionable := cp.actions[ch.Id]
		actionable.Populated = true
		actionable.CheckIn = ch

		cp.actions[ch.Id] = actionable

		log.WithFields(log.Fields{"chId": ch.Id, "entry": actionable}).Debug("Added definition to actions")
	}

}

func (cp *ChecksPreparation) Validate(version int) error {
	if !cp.VersionApplies(version) {
		return errors.New("Wrong version")
	}

	for checkId, actionable := range cp.actions {
		if actionable.Action == ActionTypeUnknown {
			return errors.New(fmt.Sprintf(
				"Check defined but not declared in manifest. check=%v", actionable.Id))
		}

		if !actionable.Populated {
			return errors.New(fmt.Sprintf(
				"Check declared in manifest but not defined. check=%v", checkId))
		}
	}

	return nil
}
