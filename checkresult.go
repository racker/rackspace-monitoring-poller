package main

import ()

var DefaultStatusLimit = 256
var DefaultStatus = "success"
var DefaultState = "available"

type States struct {
	State  string
	Status string
}

func (st *States) SetState(state string) {
	st.State = state
}

func (st *States) SetStatus(status string) {
	if len(status) > DefaultStatusLimit {
		status = status[:DefaultStatusLimit]
	}
	st.Status = status
}

type CheckResult struct {
	Metrics map[string]*Metric
}

func NewCheckResult(metrics ...*Metric) *CheckResult {
	cr := &CheckResult{
		Metrics: make(map[string]*Metric, len(metrics)+1),
	}
	cr.AddMetrics(metrics...)
	return cr
}

func (cr *CheckResult) AddMetric(metric *Metric) {
	cr.Metrics[metric.Name] = metric
}

func (cr *CheckResult) AddMetrics(metrics ...*Metric) {
	for _, metric := range metrics {
		cr.AddMetric(metric)
	}
}

func (cr *CheckResult) GetMetric(name string) *Metric {
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
	crs.SetState(DefaultState)
	crs.SetStatus(DefaultStatus)
	if cr != nil {
		crs.Add(cr)
	}
	return crs
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
