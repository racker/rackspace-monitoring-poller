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

package check

import (
	"fmt"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"strings"
)

const (
	// StateAvailable constant used for setting states and statuses
	StateAvailable = "available"
	// StateUnavailable constant used for setting states and statuses
	StateUnavailable = "unavailable"
	// StatusSuccess constant used for setting states and statuses
	StatusSuccess = "success"
	// StatusUnknownError constant used for setting states and statuses
	StatusUnknownError = "unknown error"
)

var (
	// DefaultStatusLimit sets the limit to Status string length
	DefaultStatusLimit = 256
	// DefaultStateLimit sets the limit to State string length
	DefaultStateLimit = 256
	// DefaultStatus sets the default status for a check.
	// Default is set to "success"
	DefaultStatus = StatusSuccess
	// DefaultState sets the default state for a check.
	// Default is set to "available"
	DefaultState = StateAvailable
)

// States used for check result.  Consist of State and Status
type States struct {
	State  string `json:"state"`
	Status string `json:"status"`
}

func (s States) String() string {
	return fmt.Sprintf("{state=%v, status=%v}", s.State, s.Status)
}

// SetStateAvailable updates state to available
func (st *States) SetStateAvailable() {
	st.State = StateAvailable
}

// SetStateUnavailable updates state to unavailable
func (st *States) SetStateUnavailable() {
	st.State = StateUnavailable
}

// SetStatusUnknown updates status to unknown
func (st *States) SetStatusUnknown() {
	st.Status = StatusUnknownError
}

// SetStatusSuccess updates status to success
func (st *States) SetStatusSuccess() {
	st.Status = StatusSuccess
}

// SetState sets the state of result.
// If length of state is > DefaultStateLimit, it crops it
func (st *States) SetState(state string) {
	if len(state) > DefaultStateLimit {
		state = state[:DefaultStateLimit]
	}
	st.State = state
}

// SetStatus sets the status of result.
// If length of status is > DefaultStatusLimit, it crops it
func (st *States) SetStatus(status string) {
	if len(status) > DefaultStatusLimit {
		status = status[:DefaultStatusLimit]
	}
	st.Status = status
}

func (st *States) SetStatusFromError(err error) {
	str := err.Error()
	/*
		Networking golang errors tend to be of the form
			write ip6 ::->2001:4800:7902:1:0:a:4323:44: sendto: no route to host
		so it's the last part after the colon that's meaningful and concise for check status.
	*/
	pos := strings.LastIndex(str, ":")
	if pos >= 0 {
		str = strings.TrimSpace(str[pos+1:])
	}

	st.Status = str
}

// Result wraps the metrics map
// metric name is used as key
type Result struct {
	Metrics map[string]*metric.Metric
}

func (r *Result) String() string {
	metricStrings := make(map[string]string, len(r.Metrics))
	for key, entry := range r.Metrics {
		metricStrings[key] = fmt.Sprintf("%v", entry)
	}
	return fmt.Sprintf("%v", metricStrings)
}

// NewResult creates a CheckResult object and adds passed
// in metrics.  Returns that check result
func NewResult(metrics ...*metric.Metric) *Result {
	cr := &Result{
		Metrics: make(map[string]*metric.Metric, len(metrics)+1),
	}
	cr.AddMetrics(metrics...)
	return cr
}

// AddMetric adds a new metric to the map.  If metric name is
// already in the map, it overwrites the old metric with the
// passed in
func (cr *Result) AddMetric(metric *metric.Metric) {
	cr.Metrics[metric.Name] = metric
}

// AddMetrics adds a new metrics list.
// If metric names already exist in check result, the old
// metric is overwritten with the one in the list
func (cr *Result) AddMetrics(metrics ...*metric.Metric) {
	for _, metric := range metrics {
		cr.AddMetric(metric)
	}
}

// GetMetric returns the metric from the map based on passed in
// metric name
func (cr *Result) GetMetric(name string) *metric.Metric {
	return cr.Metrics[name]
}

// ResultSet wraps the states, the check, and the CheckResult
// list.  It also provides an available boolean to show whether the
// set is available or not.  Metrics are a list of CheckResults,
// which contain a metric map (so a list of maps)
type ResultSet struct {
	States
	Check     Check     `json:"check"`
	Metrics   []*Result `json:"metrics"`
	Available bool      `json:"available"`
}

func (rs *ResultSet) String() string {
	metricStrings := make([]string, len(rs.Metrics))
	for i, result := range rs.Metrics {
		metricStrings[i] = fmt.Sprintf("%v", result)
	}

	return fmt.Sprintf("{states=%v, check=%v, metrics=[%v], available=%v}",
		rs.States, rs.Check, strings.Join(metricStrings, ","), rs.Available)
}

// NewResultSet creates a new ResultSet.
// By default the set's state is unavailable and status is unknown
// It then adds passed in check results to the set and sets the
// check to the passed in check.
func NewResultSet(ch Check, cr *Result) *ResultSet {
	crs := &ResultSet{
		Check:   ch,
		Metrics: make([]*Result, 0),
	}
	crs.SetStateUnavailable()
	crs.SetStatusUnknown()
	if cr != nil {
		crs.Add(cr)
	}
	return crs
}

// SetStateAvailable sets the set to available and sets its
// state to available
func (crs *ResultSet) SetStateAvailable() {
	crs.Available = true
	crs.States.SetStateAvailable()
}

// SetStateUnavailable sets the set to unavailable, sets its
// state to unavailable, and clears all check results
func (crs *ResultSet) SetStateUnavailable() {
	crs.Available = false
	crs.States.SetStateUnavailable()
}

// ClearMetrics clears all the check results (empties the list)
func (crs *ResultSet) ClearMetrics() {
	crs.Metrics = make([]*Result, 0)
}

// Add adds a check result to check result list
func (crs *ResultSet) Add(cr *Result) {
	crs.Metrics = append(crs.Metrics, cr)
}

// Length returns the number of check results in a set
func (crs *ResultSet) Length() int {
	return len(crs.Metrics)
}

// Get returns a check result by its index in a list.
// Stops the program if the index is out of bounds.
func (crs *ResultSet) Get(idx int) *Result {
	if idx >= crs.Length() {
		panic("CheckResultSet index is greater than length")
	}
	return crs.Metrics[idx]
}

func (crs *ResultSet) PopulateMetricsPostContent(clockOffset int64, content *protocol.MetricsPostContent) {
	content.EntityId = crs.Check.GetEntityID()
	content.CheckId = crs.Check.GetID()
	content.CheckType = crs.Check.GetCheckType()
	content.MinCheckPeriod = crs.Check.GetPeriod() * 1000
	content.State = crs.State
	content.Status = crs.Status
	content.Timestamp = utils.NowTimestampMillis() + clockOffset
	if crs.Length() == 0 {
		content.Metrics = nil
	} else {
		content.Metrics = make([]protocol.MetricWrap, crs.Length())
		for i := 0; i < crs.Length(); i++ {
			content.Metrics[i] = ConvertToMetricResults(crs.Get(i))
		}
	}

}
