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
	log "github.com/Sirupsen/logrus"
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
