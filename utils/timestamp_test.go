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

package utils_test

import (
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTimeLatencyTracking_ComputeSkew_TransitDelay(t *testing.T) {
	tracking := utils.TimeLatencyTracking{
		PollerSendTimestamp: 1000,
		ServerRecvTimestamp: 2000,
		ServerRespTimestamp: 2001,
		PollerRecvTimestamp: 3001,
	}

	offset, delay, err := tracking.ComputeSkew()

	assert.NoError(t, err)
	assert.Equal(t, int64(0), offset)
	assert.Equal(t, int64(2002), delay)
}

func TestTimeLatencyTracking_ComputeSkew_ClockOffset(t *testing.T) {
	tracking := utils.TimeLatencyTracking{
		PollerSendTimestamp: 1000,
		ServerRecvTimestamp: 3000,
		ServerRespTimestamp: 3001,
		PollerRecvTimestamp: 3001,
	}

	offset, delay, err := tracking.ComputeSkew()

	assert.NoError(t, err)
	assert.Equal(t, int64(1000), offset)
	assert.Equal(t, int64(2002), delay)
}

func TestTimeLatencyTracking_ComputeSkew_EmptyMeasurement(t *testing.T) {
	tracking := utils.TimeLatencyTracking{
		PollerSendTimestamp: 1000,
		ServerRecvTimestamp: 0,
		ServerRespTimestamp: 0,
		PollerRecvTimestamp: 3001,
	}

	_, _, err := tracking.ComputeSkew()

	assert.Error(t, err)
}
