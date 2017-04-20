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
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPinger_ValidLocalhost(t *testing.T) {
	pinger, err := check.NewPinger("test1", check.IcmpNetUDP4, "127.0.0.1")
	require.NoError(t, err)
	require.NotNil(t, pinger)
	defer pinger.Close()

	respCh := pinger.Ping(1)

	select {
	case resp := <-respCh:
		assert.Equal(t, 1, resp.Seq)
		assert.True(t, resp.Rtt > 0)
	case <-time.After(50 * time.Millisecond):
		assert.Fail(t, "Timed out waiting for response")
	}
}

func TestPinger_Invalid127(t *testing.T) {
	pinger, err := check.NewPinger("test1", check.IcmpNetUDP4, "127.0.0.2")
	require.NoError(t, err)
	require.NotNil(t, pinger)
	defer pinger.Close()

	respCh := pinger.Ping(1)

	select {
	case <-respCh:
		assert.Fail(t, "Should not have gotten a response")
	case <-time.After(10 * time.Millisecond):
	}
}

func TestPinger_ValidLocalhostIPv6(t *testing.T) {
	pinger, err := check.NewPinger("test1", check.IcmpNetUDP6, "::1")
	require.NoError(t, err)
	require.NotNil(t, pinger)
	defer pinger.Close()

	respCh := pinger.Ping(1)

	select {
	case resp := <-respCh:
		assert.Equal(t, 1, resp.Seq)
		assert.True(t, resp.Rtt > 0)
	case <-time.After(50 * time.Millisecond):
		assert.Fail(t, "Timed out waiting for response")
	}
}
