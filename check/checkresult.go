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

// Check Result
package check

import (
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
)

const (
	StateAvailable     = "available"
	StateUnavailable   = "unavailable"
	StatusSuccess      = "success"
	StatusUnknownError = "unknown error"
)

var (
	DefaultStatusLimit = 256
	DefaultStateLimit  = 256
	DefaultStatus      = StatusSuccess
	DefaultState       = StateAvailable
)

type States struct {
	State  string
	Status string
}

func (crs *States) SetStateAvailable() {
	crs.State = StateAvailable
}

func (crs *States) SetStateUnavailable() {
	crs.State = StateUnavailable
}

func (crs *States) SetStatusUnknown() {
	crs.Status = StatusUnknownError
}

func (crs *States) SetStatusSuccess() {
	crs.Status = StatusSuccess
}

func (st *States) SetState(state string) {
	if len(state) > DefaultStatusLimit {
		state = state[:DefaultStateLimit]
	}
	st.State = state
}

func (st *States) SetStatus(status string) {
	if len(status) > DefaultStatusLimit {
		status = status[:DefaultStatusLimit]
	}
	st.Status = status
}

type CheckResult struct {
	Metrics map[string]*metric.Metric
}

func NewCheckResult(metrics ...*metric.Metric) *CheckResult {
	cr := &CheckResult{
		Metrics: make(map[string]*metric.Metric, len(metrics)+1),
	}
	cr.AddMetrics(metrics...)
	return cr
}

func (cr *CheckResult) AddMetric(metric *metric.Metric) {
	cr.Metrics[metric.Name] = metric
}

func (cr *CheckResult) AddMetrics(metrics ...*metric.Metric) {
	for _, metric := range metrics {
		cr.AddMetric(metric)
	}
}

func (cr *CheckResult) GetMetric(name string) *metric.Metric {
	return cr.Metrics[name]
}

type CheckResultSet struct {
	States
	Check   Check
	Metrics []*CheckResult
}

func NewCheckResultSet(ch Check, cr *CheckResult) *CheckResultSet {
	crs := &CheckResultSet{
		Check:   ch,
		Metrics: make([]*CheckResult, 0),
	}
	crs.SetStateUnavailable()
	crs.SetStatusUnknown()
	if cr != nil {
		crs.Add(cr)
	}
	return crs
}

func (crs *CheckResultSet) SetStateUnavailable() {
	crs.States.SetStateUnavailable()
	crs.ClearMetrics()
}

func (crs *CheckResultSet) ClearMetrics() {
	crs.Metrics = make([]*CheckResult, 0)
}

func (crs *CheckResultSet) Add(cr *CheckResult) {
	crs.Metrics = append(crs.Metrics, cr)
}

func (crs *CheckResultSet) Length() int {
	return len(crs.Metrics)
}

func (crs *CheckResultSet) Get(idx int) *CheckResult {
	if idx >= crs.Length() {
		panic("CheckResultSet index is greater than length")
	}
	return crs.Metrics[idx]
}
