package main

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"time"
)

type Scheduler struct {
	ctx    context.Context
	cancel context.CancelFunc

	zoneId string
	checks map[string]Check
	input  chan Frame

	stream *ConnectionStream
}

func NewScheduler(zoneId string, stream *ConnectionStream) *Scheduler {
	s := &Scheduler{
		checks: make(map[string]Check),
		input:  make(chan Frame, 1024),
		stream: stream,
		zoneId: zoneId,
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s
}

func (s *Scheduler) Input() chan Frame {
	return s.input
}

func (s *Scheduler) Close() {
	s.cancel()
}

func (s *Scheduler) runCheck(ch Check) {
	for {
		select {
		case <-time.After(ch.GetWaitPeriod()):
			crs, err := ch.Run()
			if err != nil {
			} else {
				s.SendMetrics(crs)
			}
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Scheduler) SendMetrics(crs *CheckResultSet) {
	s.stream.SendMetrics(crs)
}

func (s *Scheduler) Register(ch Check) {
	s.checks[ch.GetId()] = ch
}

func (s *Scheduler) run() {
	for {
		select {
		case f := <-s.input:
			ch := NewCheck(f)
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
