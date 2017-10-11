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
	"testing"
	"github.com/golang/mock/gomock"
	"fmt"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/stretchr/testify/assert"
	"time"
)

// phasingEventConsumer is an event consumer that supports unit testing by providing functions for evaluating the
// existence or absence of consumed events.
type phasingEventConsumer struct {
	events chan utils.Event
}

func newPhasingEventConsumer() *phasingEventConsumer {
	return &phasingEventConsumer{
		events: make(chan utils.Event, 10),
	}
}

func (c *phasingEventConsumer) HandleEvent(evt utils.Event) error {
	c.events <- evt
	return nil
}

// waitFor will block the caller until either an event is consumed for the timeout expires.
// If an event is received, it will use the targetMatcher to evaluate the expectation of the event's Target field.
// If the timeout expires or the target field is wrong the test context will be failed.
func (c *phasingEventConsumer) waitFor(t *testing.T, timeout time.Duration, eventType string, targetMatcher gomock.Matcher) {
	select {
	case evt := <-c.events:
		assert.Equal(t, eventType, evt.Type(), "Wrong event type")
		if !targetMatcher.Matches(evt.Target()) {
			assert.Fail(t, targetMatcher.String())
		}
	case <-time.After(timeout):
		t.Fail()
		//assert.Fail(t, "Did not observe an event")
	}
}

// assertNoEvent will block the caller until timeout duration unless an event is consumed.
// If an event is consumed within timeout duration, then the test context will be failed.
func (c *phasingEventConsumer) assertNoEvent(t *testing.T, timeout time.Duration) {
	select {
	case evt := <-c.events:
		assert.Fail(t, fmt.Sprintf("Should not have seen an event, but got %v", evt.Type()))
	case <-time.After(timeout):
		break
	}
}
