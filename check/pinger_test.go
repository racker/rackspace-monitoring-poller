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
	"runtime"
	"testing"
	"time"
	"sync"
	"fmt"
	"github.com/sirupsen/logrus"
)

func TestPinger_ValidLocalhost(t *testing.T) {
	pinger, err := check.NewPinger("test1", check.IcmpNetUDP4, "127.0.0.1")
	require.NoError(t, err)
	require.NotNil(t, pinger)
	defer pinger.Close()

	resp := pinger.Ping(1, 1*time.Second)
	assert.NoError(t, resp.Err)
	assert.False(t, resp.Timeout)
	assert.Equal(t, 1, resp.Seq)
	assert.True(t, resp.Rtt > 0)
}

func TestPinger_Invalid127(t *testing.T) {
	pinger, err := check.NewPinger("test1", check.IcmpNetUDP4, "127.0.0.2")
	require.NoError(t, err)
	require.NotNil(t, pinger)
	defer pinger.Close()

	resp := pinger.Ping(1, 1*time.Second)
	assert.True(t, resp.Timeout)
	assert.Error(t, resp.Err)
}

func TestPinger_ValidLocalhostIPv6(t *testing.T) {
	pinger, err := check.NewPinger("test1", check.IcmpNetUDP6, "::1")
	require.NoError(t, err)
	require.NotNil(t, pinger)
	defer pinger.Close()

	resp := pinger.Ping(1, 1*time.Second)

	assert.NoError(t, resp.Err)
	assert.False(t, resp.Timeout)
	assert.Equal(t, 1, resp.Seq)
	assert.True(t, resp.Rtt > 0)
}

func TestPinger_Concurrent(t *testing.T) {
	if testing.Verbose() {
		logrus.SetLevel(logrus.DebugLevel)
	}

	const concurrency = 50
	const pings = 5
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(checkId string) {
			time.Sleep(10 * time.Millisecond)
			defer wg.Done()

			t.Logf("Starting %s", checkId)
			pinger, err := check.PingerFactory(checkId, "127.0.0.1", check.PingerIPv4)
			require.NoError(t, err)

			responses := make([]*check.PingResponse, pings)

			for p := 0; p < pings; p++ {
				resp := pinger.Ping(p+1, 1*time.Second)
				require.False(t, resp.Timeout, "not expecting timeout for seq=%d, checkId=%s", p+1, checkId)
				require.True(t, resp.Seq > 0 && resp.Seq <= pings, "invalid seq from resp=%v", resp)
				responses[resp.Seq-1] = &resp
				time.Sleep(10 * time.Millisecond)
			}

			for p := 0; p < pings; p++ {
				require.NotNil(t, responses[p], "Missing ping seq=%d,checkId=%s", p+1, checkId)
				assert.True(t, responses[p].Rtt > 0, "Zero RTT seq=%d", p+1)
				assert.NoError(t, responses[p].Err, "seq=%d", p+1)
				assert.False(t, responses[p].Timeout, "seq=%d", p+1)

			}

		}(fmt.Sprintf("test-%d", i))
	}

	wg.Wait()
}
