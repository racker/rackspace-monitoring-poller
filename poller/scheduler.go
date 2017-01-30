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
)

const (
	checkPreparationBufferSize = 10
)

// EleScheduler implements Scheduler interface.
// See Scheduler for more information.
type EleScheduler struct {
	ctx    context.Context
	cancel context.CancelFunc

	zoneID       string
	checks       map[string]check.Check
	preparations chan *CheckPreparation

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
		checks:       make(map[string]check.Check),
		preparations: make(chan *CheckPreparation, checkPreparationBufferSize),
		stream:       stream,
		zoneID:       zoneID,
		scheduler:    checkScheduler,
		executor:     checkExecutor,
	}

	// by default we are our own scheduler/executor of checks
	if s.scheduler == nil {
		s.scheduler = s
	}
	if s.executor == nil {
		s.executor = s
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())

	go s.runReconciler()

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

// Close cancels the context and closes the connection
func (s *EleScheduler) Close() {
	s.cancel()
}

func (s *EleScheduler) ReconcileChecks(cp *CheckPreparation) {
	s.preparations <- cp
}

func (s *EleScheduler) runReconciler() {
	for {
		select {
		case cp := <-s.preparations:

			log.WithField("cp", cp).Debug("Reconciling prepared checks")

		// TODO
		// ... compute removed checks by finding non-intersecting checks
		// ... cancel modified/removed checks
		// ... reschedule modified/added checks

		case <-s.ctx.Done():
			return
		}
	}
}

// Schedule is the default implementation of CheckScheduler that kicks off a go routine to run a check's timer.
func (s *EleScheduler) Schedule(ch check.Check) {
	go s.runCheckTimerLoop(ch)
}

func (s *EleScheduler) runCheckTimerLoop(ch check.Check) {
	// Spread the checks out over 30 seconds
	jitter := rand.Intn(CheckSpreadInMilliseconds) + 1

	log.WithFields(log.Fields{
		"check":      ch.GetID(),
		"jitterMs":   jitter,
		"waitPeriod": ch.GetWaitPeriod(),
	}).Info("Starting check")

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
}

// Execute perform the default CheckExecutor behavior by running the check and sending its results via SendMetrics.
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
