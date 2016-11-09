package main

import (
	"encoding/json"
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
	Id             string             `json:"id"`
	CheckType      string             `json:"type"`
	Period         uint64             `json:"period"`
	Timeout        uint64             `json:"timeout"`
	EntityId       string             `json:"entity_id"`
	ZoneId         string             `json:"zone_id"`
	Details        *json.RawMessage   `json:"details"`
	Disabled       bool               `json:"disabled"`
	IpAddresses    *map[string]string `json:"ip_addresses"`
	TargetAlias    *string            `json:"target_alias"`
	TargetHostname *string            `json:"target_hostname"`
	TargetResolver *string            `json:"target_resolver"`
}

func NewCheck(f Frame) Check {
	checkBase := &CheckBase{}
	err := json.Unmarshal(*f.GetRawParams(), &checkBase)
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
	log.WithFields(log.Fields{
		"type":    ch.GetCheckType(),
		"period":  ch.GetPeriod(),
		"timeout": ch.GetTimeout(),
		"details": string(*ch.Details),
	}).Infof("New check %v", ch.GetId())
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
	ch.Timeout = timeout
}

func (ch *CheckBase) GetWaitPeriod() time.Duration {
	return time.Duration(ch.Period) * time.Second
}
