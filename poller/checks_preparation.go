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
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/racker/rackspace-monitoring-poller/protocol"
	"github.com/racker/rackspace-monitoring-poller/protocol/check"
)

// ActionType provides a non-protocol enumeration of valid check preparation actions
type ActionType int

const (
	ActionTypeUnknown = iota
	ActionTypeStart
	ActionTypeRestart
	ActionTypeContinue
)

// ActionableCheck enriches a check.CheckIn with an action indicator.
// It is expected that during preparation these may have only abbreviated
// fields populated in check.CheckIn.
type ActionableCheck struct {
	check.CheckIn

	Action ActionType `json:"action"`
	// Populated indicates if the check.CheckIn has been fully populated; however,
	// this is not applicable if Action is protocol.PrepareActionContinue.
	Populated bool `json:"populated"`
}

func (ac ActionableCheck) String() string {
	json, _ := json.Marshal(ac)
	return string(json)
}

// ChecksPreparing conveys ActionableCheck instances are are ready to be validated.
type ChecksPreparing interface {
	GetActionableChecks() (actionableChecks []*ActionableCheck)
}

// ChecksPrepared conveys ActionableCheck instances that are fully populated and ready to be reconciled
type ChecksPrepared interface {
	GetActionableChecks() (actionableChecks []*ActionableCheck)
	Stringer() string
}

type ChecksPreparation struct {
	ZoneId  string `json:"zone_id"`
	Version int    `json:"version"`

	// Actions is a map of checkId->ActionableCheck
	Actions map[string] /*checkId*/ *ActionableCheck `json:"actions"`
}

// NewChecksPreparation initiates a new checks preparation session.
// Returns the new ChecksPreparation if successful or an error if an unsupported action type was encountered.
func NewChecksPreparation(zoneId string, version int, manifest []protocol.PollerPrepareManifest) (*ChecksPreparation, error) {
	cp := &ChecksPreparation{
		ZoneId:  zoneId,
		Version: version,
		Actions: make(map[string]*ActionableCheck),
	}

	for _, m := range manifest {

		actionType := mapToActionType(m.Action)
		if actionType == ActionTypeUnknown {
			return nil, errors.New(fmt.Sprintf("Unsupported action in manifest: action=%s, check=%s", m.Action, m.Id))
		}
		cp.Actions[m.Id] = &ActionableCheck{
			Action: actionType,
			// "pre populate" actions we don't expect to see defined
			Populated: !doesCheckPreparationNeedPopulating(m.Action),

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
	log.WithFields(log.Fields{
		"prefix":  cp.GetLogPrefix(),
		"actions": cp.Actions,
	}).Debug("Prepared checks from manifest")
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

func doesCheckPreparationNeedPopulating(action string) bool {
	return action != protocol.PrepareActionContinue
}

func (cp *ChecksPreparation) Stringer() string {
	bytes, err := json.Marshal(cp)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func (cp *ChecksPreparation) GetLogPrefix() string {
	return fmt.Sprintf("checks.preparation zoneid=%v, version=%v", cp.ZoneId, cp.Version)
}

func (cp *ChecksPreparation) GetActionableChecks() (actionableChecks []*ActionableCheck) {
	actionableChecks = make([]*ActionableCheck, 0, len(cp.Actions))
	for _, ac := range cp.Actions {
		actionableChecks = append(actionableChecks, ac)
	}
	return
}

func (cp *ChecksPreparation) VersionApplies(version int) bool {
	return cp != nil && cp.Version == version
}

func (cp *ChecksPreparation) IsNewer(version int) bool {
	return cp == nil || version > cp.Version
}

func (cp *ChecksPreparation) IsOlder(version int) bool {
	return cp != nil && version < cp.Version
}

func (cp *ChecksPreparation) AddDefinitions(block []*check.CheckIn) {
	for _, ch := range block {
		actionable, ok := cp.Actions[ch.Id]
		if !ok {
			// place "undefined" entry for validation
			cp.Actions[ch.Id] = &ActionableCheck{}
			log.WithFields(log.Fields{
				"prefix":   cp.GetLogPrefix(),
				"check_id": ch.Id,
			}).Warn("Trying to add definition for check that wasn't previously declared")
			continue
		}
		actionable.Populated = true
		actionable.CheckIn = *ch

		if log.GetLevel() >= log.DebugLevel {
			log.WithFields(log.Fields{
				"prefix":   cp.GetLogPrefix(),
				"check_id": ch.Id,
				"entry":    actionable.String(),
			}).Debug("Added definition to actions")
		}
	}
}

func (cp *ChecksPreparation) Validate(version int) error {
	if !cp.VersionApplies(version) {
		return errors.New("Wrong version")
	}

	for checkId, actionable := range cp.Actions {
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

func (cp *ChecksPreparation) GetPreparationZoneId() string {
	return cp.ZoneId
}

func (cp *ChecksPreparation) GetPreparationVersion() int {
	return cp.Version
}
