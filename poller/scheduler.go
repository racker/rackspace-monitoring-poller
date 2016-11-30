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
	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/check"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"math/rand"
	"time"
)

const (
	CheckSpreadInMilliseconds = 30000
)

type Scheduler struct {
	ctx    context.Context
	cancel context.CancelFunc

	zoneId string
	checks map[string]check.Check
	input  chan protocol.Frame

	stream *ConnectionStream
}

func NewScheduler(zoneId string, stream *ConnectionStream) *Scheduler {
	s := &Scheduler{
		checks: make(map[string]check.Check),
		input:  make(chan protocol.Frame, 1024),
		stream: stream,
		zoneId: zoneId,
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s
}

func (s *Scheduler) Input() chan protocol.Frame {
	return s.input
}

func (s *Scheduler) Close() {
	s.cancel()
}

func (s *Scheduler) runCheck(ch check.Check) {
	// Spread the checks out over 30 seconds
	jitter := rand.Intn(CheckSpreadInMilliseconds) + 1
	time.Sleep(time.Duration(jitter) * time.Millisecond)
	for {
		select {
		case <-time.After(ch.GetWaitPeriod()):
			crs, err := ch.Run()
			if err != nil {
				log.Errorf("Error running check: %v", err)
			} else {
				s.SendMetrics(crs)
			}
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Scheduler) SendMetrics(crs *check.CheckResultSet) {
	s.stream.SendMetrics(crs)
}

func (s *Scheduler) Register(ch check.Check) {
	s.checks[ch.GetId()] = ch
}

func (s *Scheduler) run() {
	for {
		select {
		case f := <-s.input:
			ch := check.NewCheck(f.GetRawParams())
			if ch == nil {
				log.Printf("Invalid Check")
				continue
			}
			s.Register(ch)
			go s.runCheck(ch)
		case <-s.ctx.Done():
			return
		}
	}
}
