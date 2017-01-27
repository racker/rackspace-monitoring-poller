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

package poller_test

import (
	"github.com/racker/rackspace-monitoring-poller/poller"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"github.com/racker/rackspace-monitoring-poller/protocol/check"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewCheckPreparation_VersionMatch(t *testing.T) {
	cp := poller.NewCheckPreparation(1, []protocol.PollerPrepareManifest{})

	assert.NotNil(t, cp)
	assert.True(t, cp.VersionApplies(1))
}

func TestNewCheckPreparation_VersionMismatch(t *testing.T) {
	cp := poller.NewCheckPreparation(1, []protocol.PollerPrepareManifest{})

	assert.NotNil(t, cp)
	assert.False(t, cp.VersionApplies(3))
}

func TestCheckPreparation_AddDefinitions_Normal(t *testing.T) {
	manifest := []protocol.PollerPrepareManifest{
		{
			Action:   protocol.PrepareActionStart,
			ZoneId:   "zn1",
			EntityId: "en1",
			Id:       "ch1",
		},
		{
			Action:   protocol.PrepareActionRestart,
			ZoneId:   "zn1",
			EntityId: "en2",
			Id:       "ch2",
		},
		{
			Action:   protocol.PrepareActionContinue,
			ZoneId:   "zn1",
			EntityId: "en2",
			Id:       "ch3",
		},
	}

	cp := poller.NewCheckPreparation(1, manifest)

	block1 := []check.CheckIn{
		{
			CheckHeader: check.CheckHeader{
				Id:        "ch2",
				ZoneId:    "zn1",
				EntityId:  "en2",
				Period:    60,
				CheckType: "remote.tcp",
			},
		},
	}

	block2 := []check.CheckIn{
		{
			CheckHeader: check.CheckHeader{
				Id:        "ch1",
				ZoneId:    "zn1",
				EntityId:  "en1",
				Period:    60,
				CheckType: "remote.tcp",
			},
		},
	}

	cp.AddDefinitions(block1)
	cp.AddDefinitions(block2)

	err := cp.Validate(1)
	assert.NoError(t, err)
}

func TestCheckPreparation_AddDefinitions_MissingOne(t *testing.T) {
	manifest := []protocol.PollerPrepareManifest{
		{
			Action:   protocol.PrepareActionStart,
			ZoneId:   "zn1",
			EntityId: "en1",
			Id:       "ch1",
		},
		{
			Action:   protocol.PrepareActionRestart,
			ZoneId:   "zn1",
			EntityId: "en2",
			Id:       "ch2",
		},
		{
			Action:   protocol.PrepareActionContinue,
			ZoneId:   "zn1",
			EntityId: "en2",
			Id:       "ch3",
		},
	}

	cp := poller.NewCheckPreparation(1, manifest)

	block1 := []check.CheckIn{
		{
			CheckHeader: check.CheckHeader{
				Id:        "ch2",
				ZoneId:    "zn1",
				EntityId:  "en2",
				Period:    60,
				CheckType: "remote.tcp",
			},
		},
	}

	cp.AddDefinitions(block1)

	err := cp.Validate(1)
	assert.Error(t, err)

}

func TestCheckPreparation_AddDefinitions_WrongVersion(t *testing.T) {
	manifest := []protocol.PollerPrepareManifest{
		{
			Action:   protocol.PrepareActionRestart,
			ZoneId:   "zn1",
			EntityId: "en2",
			Id:       "ch2",
		},
	}

	cp := poller.NewCheckPreparation(1, manifest)

	block1 := []check.CheckIn{
		{
			CheckHeader: check.CheckHeader{
				Id:        "ch2",
				ZoneId:    "zn1",
				EntityId:  "en2",
				Period:    60,
				CheckType: "remote.tcp",
			},
		},
	}

	cp.AddDefinitions(block1)

	err := cp.Validate(2)
	assert.Error(t, err)

}
