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

// The check package contains an implementation file per check-type.
//
// A typical/starter check implementation file would contain something like:
//
//  type <TYPE>Check struct {
//    CheckBase
//    Details struct {
//      SomeField string|... `json:"some_field"`
//      ...
//    }
//  }
//
//  func New<TYPE>Check(base *CheckBase) Check {
//    check := &<TYPE>Check{CheckBase: *base}
//    err := json.Unmarshal(*base.Details, &check.Details)
//    if err != nil {
//      log.Printf("Error unmarshalling check details")
//      return nil
//    }
//    check.PrintDefaults()
//    return check
//  }
//
//  func (ch *<TYPE>Check) Run() (*CheckResultSet, error) {
//    log.Printf("Running <TYPE> Check: %v", ch.GetId())
//    ...do check specifics...
//
//    ...upon success:
//    cr := NewCheckResult(
//      metric.NewMetric(...)
//      ...
//    )
//    crs := NewCheckResultSet(ch, cr)
//    crs.SetStateAvailable()
//    return crs, nil
//  }
package check

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"time"
)

type Check interface {
	GetId() string
	SetId(id string)
	GetEntityId() string
	GetCheckType() string
	SetCheckType(checkType string)
	GetPeriod() uint64
	GetWaitPeriod() time.Duration
	SetPeriod(period uint64)
	GetTimeout() uint64
	SetTimeout(timeout uint64)
	Run() (*CheckResultSet, error)
}

type CheckBase struct {
	Id             string            `json:"id"`
	CheckType      string            `json:"type"`
	Period         uint64            `json:"period"`
	Timeout        uint64            `json:"timeout"`
	EntityId       string            `json:"entity_id"`
	ZoneId         string            `json:"zone_id"`
	Details        *json.RawMessage  `json:"details"`
	Disabled       bool              `json:"disabled"`
	IpAddresses    map[string]string `json:"ip_addresses"`
	TargetAlias    *string           `json:"target_alias"`
	TargetHostname *string           `json:"target_hostname"`
	TargetResolver *string           `json:"target_resolver"`
}

func NewCheck(rawParams json.RawMessage) Check {
	checkBase := &CheckBase{}
	err := json.Unmarshal(rawParams, &checkBase)
	if err != nil {
		log.Printf("Error unmarshalling checkbase")
		return nil
	}
	switch checkBase.CheckType {
	case "remote.tcp":
		return NewTCPCheck(checkBase)
	case "remote.http":
		return NewHTTPCheck(checkBase)
	case "remote.ping":
		return NewPingCheck(checkBase)
	default:

		log.Printf("Invalid check type: %v", checkBase.CheckType)
	}
	return nil
}

func NewMetricsPostRequest(crs *CheckResultSet) *protocol.MetricsPostRequest {
	req := &protocol.MetricsPostRequest{}
	req.Version = "1"
	req.Method = "check_metrics.post"
	req.Params.EntityId = crs.Check.GetEntityId()
	req.Params.CheckId = crs.Check.GetId()
	req.Params.CheckType = crs.Check.GetCheckType()
	req.Params.Metrics = []protocol.MetricWrap{ConvertToMetricResults(crs)}
	req.Params.MinCheckPeriod = crs.Check.GetPeriod() * 1000
	req.Params.State = crs.State
	req.Params.Status = crs.Status
	req.Params.Timestamp = utils.NowTimestampMillis()
	return req
}

func ConvertToMetricResults(crs *CheckResultSet) protocol.MetricWrap {
	wrappers := make(protocol.MetricWrap, 0)
	wrappers = append(wrappers, nil) // needed for the current protocol
	for i := 0; i < crs.Length(); i++ {
		cr := crs.Get(i)
		mapper := make(map[string]*protocol.MetricTVU)
		for key, m := range cr.Metrics {
			mapper[key] = &protocol.MetricTVU{
				Type:  m.TypeString,
				Value: fmt.Sprintf("%v", m.Value),
				Unit:  m.Unit,
			}
		}
		wrappers = append(wrappers, mapper)
	}
	return wrappers
}

func (ch *CheckBase) PrintDefaults() {
	var targetAlias string
	var targetHostname string
	var targetResolver string
	if ch.TargetAlias != nil {
		targetAlias = *ch.TargetAlias
	}
	if ch.TargetHostname != nil {
		targetHostname = *ch.TargetHostname
	}
	if ch.TargetResolver != nil {
		targetResolver = *ch.TargetResolver
	}
	log.WithFields(log.Fields{
		"type":            ch.CheckType,
		"period":          ch.Period,
		"timeout":         ch.Timeout,
		"disabled":        ch.Disabled,
		"ipaddresses":     ch.IpAddresses,
		"target_alias":    targetAlias,
		"target_hostname": targetHostname,
		"target_resolver": targetResolver,
		"details":         string(*ch.Details),
	}).Infof("New check %v", ch.GetId())
}

// GetTargetIP obtains the specific IP address selected for this check.
// It returns the resolved IP address as dotted string.
func (ch *CheckBase) GetTargetIP() (string, error) {
	ip, ok := ch.IpAddresses[*ch.TargetAlias]
	if ok {
		return ip, nil
	}
	return "", errors.New("Invalid Target IP")

}

func (ch *CheckBase) GetId() string {
	return ch.Id
}

func (ch *CheckBase) SetId(id string) {
	ch.Id = id
}

func (ch *CheckBase) GetCheckType() string {
	return ch.CheckType
}

func (ch *CheckBase) SetCheckType(checkType string) {
	ch.CheckType = checkType
}

func (ch *CheckBase) GetPeriod() uint64 {
	return ch.Period
}

func (ch *CheckBase) GetEntityId() string {
	return ch.EntityId
}

func (ch *CheckBase) SetPeriod(period uint64) {
	ch.Period = period
}

func (ch *CheckBase) GetTimeout() uint64 {
	return ch.Timeout
}

func (ch *CheckBase) SetTimeout(timeout uint64) {

}

func (ch *CheckBase) GetWaitPeriod() time.Duration {
	return time.Duration(ch.Period) * time.Second
}
