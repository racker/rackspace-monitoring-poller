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

package check_test

import (
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCheckBase_GetTargetIP(t *testing.T) {

	cb := &check.CheckBase{}

	cb.IpAddresses = map[string]string{
		"host1": "192.168.0.1",
	}
	targetAlias := "host1"
	cb.TargetAlias = &targetAlias

	result, err := cb.GetTargetIP()
	assert.NoError(t, err)
	assert.Equal(t, "192.168.0.1", result)
}

func TestCheckBase_GetTargetIP_mismatch(t *testing.T) {

	cb := &check.CheckBase{}

	cb.IpAddresses = map[string]string{
		"host1": "192.168.0.1",
	}
	targetAlias := "totallyWrongHost"
	cb.TargetAlias = &targetAlias

	_, err := cb.GetTargetIP()
	assert.Error(t, err)
}

func TestCheckBase_GetWaitPeriod(t *testing.T) {
	cb := &check.CheckBase{}

	cb.Period = 5

	assert.Equal(t, 5*time.Second, cb.GetWaitPeriod())
}
