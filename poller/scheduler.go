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

package poller

import (
	"context"
	"math/rand"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/protocol"
)

// EleScheduler implements Scheduler interface.
// See Scheduler for more information.
type EleScheduler struct {
	ctx    context.Context
	cancel context.CancelFunc

	zoneID string
	checks map[string]check.Check
	input  chan protocol.Frame

	stream ConnectionStream

	scheduler CheckScheduler
	executor  CheckExecutor
}

// NewScheduler instantiates a new Scheduler with standard scheduling and executor behaviors.
// It sets up checks, context, and passed in zoneid
func NewScheduler(zoneID string, stream ConnectionStream) Scheduler {
	return NewCustomScheduler(zoneID, stream, nil, nil)
}

// NewCustomScheduler instantiates a new Scheduler using NewScheduler but allows for more customization.
// Nil can be passed to either checkScheduler and/or checkExecutor to enable the default behavior.
func NewCustomScheduler(zoneID string, stream ConnectionStream, checkScheduler CheckScheduler, checkExecutor CheckExecutor) Scheduler {
	s := &EleScheduler{
		checks:    make(map[string]check.Check),
		input:     make(chan protocol.Frame, 1024),
		stream:    stream,
		zoneID:    zoneID,
		scheduler: checkScheduler,
		executor:  checkExecutor,
	}

	// by default we are our own scheduler/executor of checks
	if s.scheduler == nil {
		s.scheduler = s
	}
	if s.executor == nil {
		s.executor = s
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())

	return s

}

// GetZoneID retrieves zone id
func (s *EleScheduler) GetZoneID() string {
	return s.zoneID
}

// GetContext retrieves cancelable context
func (s *EleScheduler) GetContext() (ctx context.Context, cancel context.CancelFunc) {
	return s.ctx, s.cancel
}

// GetChecks retrieves check map
func (s *EleScheduler) GetChecks() map[string]check.Check {
	return s.checks
}

// GetInput returns protocol.Frame channel
func (s *EleScheduler) GetInput() chan protocol.Frame {
	return s.input
}

// Close cancels the context and closes the connection
func (s *EleScheduler) Close() {
	s.cancel()
}

func (s *EleScheduler) Schedule(ch check.Check) {
	// Spread the checks out over 30 seconds
	jitter := rand.Intn(CheckSpreadInMilliseconds) + 1

	log.WithFields(log.Fields{
		"check":      ch.GetID(),
		"jitterMs":   jitter,
		"waitPeriod": ch.GetWaitPeriod(),
	}).Info("Starting check")

	go func() {
		time.Sleep(time.Duration(jitter) * time.Millisecond)
		for {
			select {
			case <-time.After(ch.GetWaitPeriod()):
				s.executor.Execute(ch)

			case <-ch.Done(): // session cancellation is propagated since check context is child of session context
				log.WithField("check", ch.GetID()).Info("Check or session has been cancelled")
				return
			}
		}

	}()
}

func (s *EleScheduler) Execute(ch check.Check) {
	crs, err := ch.Run()
	if err != nil {
		log.Errorf("Error running check: %v", err)
	} else {
		s.SendMetrics(crs)
	}

}

// SendMetrics sends metrics passed in crs parameter via the stream
func (s *EleScheduler) SendMetrics(crs *check.ResultSet) {
	s.stream.SendMetrics(crs)
}

// Register registers the passed in ch check in the checks list
func (s *EleScheduler) Register(ch check.Check) error {
	if ch == nil {
		return ErrCheckEmpty
	}
	s.checks[ch.GetID()] = ch
	return nil
}

// RunFrameConsumer method runs the check.  It sets up the new check
// and sends it.
func (s *EleScheduler) RunFrameConsumer() {
	for {
		select {
		case f := <-s.GetInput():

			// TODO Later this will probably need to handle check cancellations in which case it can call ch.Cancel()

			checkCtx, cancelFunc := context.WithCancel(s.ctx)
			ch := check.NewCheck(checkCtx, f.GetRawParams(), cancelFunc)
			if ch == nil {
				log.Printf("Invalid Check")
				continue
			}
			s.Register(ch)
			s.scheduler.Schedule(ch)
		case <-s.ctx.Done():
			return
		}
	}
}
